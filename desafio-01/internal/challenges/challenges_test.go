package challenges

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProgressValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "progress_test_*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	p := NewProgress(tempDir)

	_, ok, _, _ := p.SubmitFlag("FLAG{ERRADA}")
	if ok {
		t.Errorf("esperava falso para flag errada")
	}

	id, ok, already, _ := p.SubmitFlag(Flag1)
	if !ok || id != 1 || already {
		t.Errorf("esperava sucesso para Flag1, obteve id=%d, ok=%v, already=%v", id, ok, already)
	}

	_, ok2, already2, _ := p.SubmitFlag(Flag1)
	if !ok2 || !already2 {
		t.Errorf("esperava flag já completada na segunda vez")
	}

	p.SubmitFlag(Flag2)
	p.SubmitFlag(Flag3)
	p.SubmitFlag(Flag4)

	if !p.IsAllCompleted() {
		t.Errorf("esperava todas completadas")
	}
	if p.TotalCompleted() != 4 {
		t.Errorf("esperava 4 completadas, obteve %d", p.TotalCompleted())
	}
}

func TestChallenge1Setup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "chal1_test_*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := SetupChallenge1(tempDir); err != nil {
		t.Fatalf("SetupChallenge1 falhou: %v", err)
	}

	var passFiles []string
	_ = filepath.Walk(filepath.Join(tempDir, "wclv_archive"), func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == "pass.txt" {
			passFiles = append(passFiles, path)
		}
		return nil
	})

	if len(passFiles) < 10 {
		t.Errorf("esperava pelo menos 10 arquivos pass.txt (incluindo decoys), encontrou %d", len(passFiles))
	}

	flagsFound := 0
	for _, p := range passFiles {
		data, _ := os.ReadFile(p)
		if strings.Contains(string(data), Flag1) {
			flagsFound++
		}
	}

	if flagsFound != 1 {
		t.Errorf("esperava exatamente 1 arquivo contendo a flag real, encontrou %d", flagsFound)
	}

	// Verifica se os arquivos de lore de Local 58 estão presentes
	bulletin := filepath.Join(tempDir, "wclv_archive", "broadcast_tapes", "weather_service", "eas_bulletins", "bulletin_01.txt")
	if _, err := os.Stat(bulletin); err != nil {
		t.Errorf("bulletin_01.txt não encontrado: %v", err)
	}
}

func TestChallenge2ServerMultiStep(t *testing.T) {
	srv := NewServerChallenge2(8099)
	if err := srv.Start(); err != nil {
		t.Fatalf("falha ao iniciar srv2: %v", err)
	}
	defer srv.Stop()

	time.Sleep(100 * time.Millisecond)

	// Step 1: Redirecionamento 302 em /weather-service -> /skywatching
	clientNoRedirect := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := clientNoRedirect.Get("http://localhost:8099/weather-service")
	if err != nil {
		t.Fatalf("GET /weather-service falhou: %v", err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Errorf("GET /weather-service deveria retornar 302, retornou %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/skywatching" {
		t.Errorf("Location deveria ser /skywatching, obteve %s", loc)
	}
	resp.Body.Close()

	// Step 2: /skywatching sem parâmetros -> 400
	resp, err = http.Get("http://localhost:8099/skywatching")
	if err != nil {
		t.Fatalf("GET /skywatching falhou: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GET /skywatching sem params deveria ser 400, obteve %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 2 com query param (?sky=visible) e header (X-Transmission: 476MHz) -> 200
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8099/skywatching?sky=visible", nil)
	req.Header.Set("X-Transmission", "476MHz")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET skywatching com headers falhou: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET skywatching com headers deveria ser 200, obteve %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 3: /rejoice sem basic auth -> 401
	payload := `{"action": "LOOK_AT_THE_MOON"}`
	req, _ = http.NewRequest(http.MethodPost, "http://localhost:8099/rejoice", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST rejoice sem auth falhou: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("POST rejoice sem auth deveria ser 401, obteve %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 4: POST /rejoice com auth (viewer:moon) e payload correto -> 200 e Flag 2
	req, _ = http.NewRequest(http.MethodPost, "http://localhost:8099/rejoice", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("viewer", "moon")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST rejoice falhou: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("POST rejoice deveria ser 200, obteve %d", resp.StatusCode)
	}
	var resBody map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&resBody)
	resp.Body.Close()

	if resBody["flag"] != Flag2 {
		t.Errorf("esperava flag %s, obteve %v", Flag2, resBody["flag"])
	}

	if srv.GetStepReached() != 4 {
		t.Errorf("esperava stepReached == 4, obteve %d", srv.GetStepReached())
	}
}

func TestChallenge3GitBranchSetup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "chal3_test_*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := SetupChallenge3(tempDir); err != nil {
		t.Fatalf("SetupChallenge3 falhou: %v", err)
	}

	repoDir := filepath.Join(tempDir, "wclv_station_repo")

	// 1. Confere que a branch atual é 'main'
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current falhou: %v", err)
	}
	if strings.TrimSpace(string(out)) != "main" {
		t.Errorf("branch inicial deveria ser main, obteve %s", string(out))
	}

	// 2. Confere que a Flag3 NÃO está na main
	cmd = exec.Command("git", "log", "-p")
	cmd.Dir = repoDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log falhou: %v", err)
	}
	if strings.Contains(string(out), Flag3) {
		t.Errorf("Flag3 não deveria estar presente no histórico da branch main!")
	}

	// 3. Confere que a branch 'weather-service-hijack' existe
	cmd = exec.Command("git", "branch", "-a")
	cmd.Dir = repoDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git branch -a falhou: %v", err)
	}
	if !strings.Contains(string(out), "weather-service-hijack") {
		t.Errorf("branch weather-service-hijack deveria existir, obteve:\n%s", string(out))
	}

	// 4. Confere que a Flag3 é recuperável na branch weather-service-hijack
	cmd = exec.Command("git", "log", "-p", "weather-service-hijack")
	cmd.Dir = repoDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log na hijack branch falhou: %v", err)
	}
	if !strings.Contains(string(out), Flag3) {
		t.Errorf("Flag3 deveria estar presente no git log da branch weather-service-hijack!")
	}
}

func TestProgressAdvancedValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "progress_adv_*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	p := NewProgress(tempDir)

	// Teste de entrada vazia
	res := p.Submit("   ")
	if res.Kind != SubmitEmpty {
		t.Errorf("esperava SubmitEmpty, obteve %v", res.Kind)
	}

	// Teste de decoy (flag falsa)
	res = p.Submit("FAKE_FLAG{LOOK_AT_THE_MOON_NOW}")
	if res.Kind != SubmitDecoy {
		t.Errorf("esperava SubmitDecoy, obteve %v", res.Kind)
	}

	// Teste de formato inválido
	res = p.Submit("chaveteste123")
	if res.Kind != SubmitBadFormat {
		t.Errorf("esperava SubmitBadFormat, obteve %v", res.Kind)
	}

	// Teste de lowercase (WrongCase)
	res = p.Submit("flag{his_throne_rejoice_look_at_the_moon}")
	if res.Kind != SubmitWrongCase {
		t.Errorf("esperava SubmitWrongCase, obteve %v", res.Kind)
	}

	// Teste de colagem suja (ex: trecho de JSON com aspas e vírgula)
	dirtyPaste := `"flag": "` + Flag1 + `",`
	res = p.Submit(dirtyPaste)
	if res.Kind != SubmitCorrect || res.ChallengeID != 1 {
		t.Errorf("esperava aceitar colagem suja com Flag1, obteve %v", res)
	}

	// Teste de submissão duplicada
	res = p.Submit(Flag1)
	if res.Kind != SubmitAlready {
		t.Errorf("esperava SubmitAlready para Flag1 duplicada, obteve %v", res.Kind)
	}

	// Teste de dicas progressivas
	if p.HintsUsed(1) != 0 {
		t.Errorf("esperava 0 dicas usadas inicialmente")
	}
	used := p.RevealHint(1, 3)
	if used != 1 || p.HintsUsed(1) != 1 {
		t.Errorf("esperava 1 dica usada após reveal")
	}
	p.RevealHint(1, 3)
	p.RevealHint(1, 3)
	usedMax := p.RevealHint(1, 3)
	if usedMax != 3 || p.HintsUsed(1) != 3 {
		t.Errorf("esperava travar no máximo de 3 dicas")
	}
	if p.TotalHintsUsed() != 3 {
		t.Errorf("esperava total de 3 dicas usadas, obteve %d", p.TotalHintsUsed())
	}
}

func TestChallenge2HeadersAndTelemetry(t *testing.T) {
	srv := NewServerChallenge2(8098)
	if err := srv.Start(); err != nil {
		t.Fatalf("falha ao iniciar srv: %v", err)
	}
	defer srv.Stop()

	// 1. Rota raiz: verifica cabeçalhos da estação WCLV
	resp, err := http.Get("http://localhost:8098/")
	if err != nil {
		t.Fatalf("GET / falhou: %v", err)
	}
	if resp.Header.Get("Server") != "WCLV-TV/1934" {
		t.Errorf("cabeçalho Server incorreto: %s", resp.Header.Get("Server"))
	}
	if resp.Header.Get("X-Station") != "LOCAL 58 (WCLV-TV)" {
		t.Errorf("cabeçalho X-Station incorreto: %s", resp.Header.Get("X-Station"))
	}
	if resp.Header.Get("X-Signal-Step") == "" {
		t.Errorf("cabeçalho X-Signal-Step ausente")
	}
	resp.Body.Close()

	// 2. Método incorreto em /skywatching -> 405 com cabeçalho Allow: GET
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8098/skywatching", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /skywatching falhou: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("esperava 405 para POST em skywatching, obteve %d", resp.StatusCode)
	}
	if resp.Header.Get("Allow") != "GET" {
		t.Errorf("esperava Allow: GET, obteve %s", resp.Header.Get("Allow"))
	}
	resp.Body.Close()

	// 3. Verifica se os eventos foram publicados no canal Events
	select {
	case evt := <-srv.Events:
		if evt.Path != "/" && evt.Path != "/skywatching" {
			t.Errorf("evento HTTP inesperado: %+v", evt)
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("nenhum evento de telemetria recebido no canal srv.Events")
	}
}

func TestChallenge4WriteFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "chal4_test_*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := WriteChallenge4Files(tempDir); err != nil {
		t.Fatalf("WriteChallenge4Files falhou: %v", err)
	}

	df := filepath.Join(Challenge4Dir(tempDir), "Dockerfile")
	content, err := os.ReadFile(df)
	if err != nil {
		t.Fatalf("falha ao ler Dockerfile gerado: %v", err)
	}

	if !strings.Contains(string(content), Flag4) {
		t.Errorf("Dockerfile deve conter a Flag4 no ENV PRESERVATION_OVERRIDE_CODE")
	}
	if !strings.Contains(string(content), `ENV PRESERVATION_OVERRIDE_CODE=""`) {
		t.Errorf("Dockerfile deve sobrescrever a variável na camada seguinte")
	}
}

