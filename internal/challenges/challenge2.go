package challenges

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPLogEvent é a telemetria de uma requisição recebida pelo transmissor.
// É consumida pela TUI para montar o log ao vivo e o inspetor de requisições.
type HTTPLogEvent struct {
	Timestamp  time.Time
	Method     string
	Path       string
	Query      string
	StatusCode int
	Message    string // diagnóstico do servidor (o que aconteceu)
	Hint       string // próximo passo sugerido (o que fazer)
	Success    bool
	Step       int // estágio do desafio concluído por esta requisição (0 = nenhum)
	RemoteAddr string
	UserAgent  string
	Headers    map[string]string // cabeçalhos relevantes recebidos (Authorization mascarado)
	Body       string            // prévia do corpo recebido
	Duration   time.Duration
	Bytes      int
}

// Target devolve "path?query" para exibição.
func (e HTTPLogEvent) Target() string {
	if e.Query == "" {
		return e.Path
	}
	return e.Path + "?" + e.Query
}

// Estágios do desafio 2 (foco telescópico).
const (
	StepRedirect = 1 // seguiu o 302 de /weather-service
	StepTuned    = 2 // query + header corretos em /skywatching
	StepAuth     = 3 // Basic Auth aceita em /rejoice
	StepRejoice  = 4 // POST JSON correto -> flag liberada
	TotalSteps   = 4
)

const maxBodyBytes = 64 << 10

type ServerChallenge2 struct {
	server      *http.Server
	listener    net.Listener
	Events      chan HTTPLogEvent
	mu          sync.Mutex
	isRunning   bool
	port        int
	StepReached int
}

func NewServerChallenge2(port int) *ServerChallenge2 {
	if port <= 0 {
		port = 8058
	}
	return &ServerChallenge2{
		port:   port,
		Events: make(chan HTTPLogEvent, 100),
	}
}

// Port retorna a porta configurada do transmissor.
func (s *ServerChallenge2) Port() int { return s.port }

// URL retorna a URL base usada nas dicas e respostas.
func (s *ServerChallenge2) URL() string { return fmt.Sprintf("http://localhost:%d", s.port) }

// Running informa se o servidor está aceitando conexões.
func (s *ServerChallenge2) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

// ---------------------------------------------------------------------------
// Middleware de telemetria
// ---------------------------------------------------------------------------

type ctxKey struct{}

// outcome é preenchido pelos handlers e lido pelo middleware para gerar o log.
type outcome struct {
	Message string
	Hint    string
	Success bool
	Step    int
	Body    string
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func outcomeFrom(r *http.Request) *outcome {
	if o, ok := r.Context().Value(ctxKey{}).(*outcome); ok {
		return o
	}
	return &outcome{}
}

// instrument envolve todas as rotas: adiciona cabeçalhos da emissora e emite
// exatamente um HTTPLogEvent por requisição (inclusive 404/405).
func (s *ServerChallenge2) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		o := &outcome{}
		rec := &statusRecorder{ResponseWriter: w}

		h := rec.Header()
		h.Set("Server", "WCLV-TV/1934")
		h.Set("X-Station", "LOCAL 58 (WCLV-TV)")
		h.Set("X-Carrier", "UHF 476.000 MHz")

		next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), ctxKey{}, o)))

		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		s.emit(HTTPLogEvent{
			Timestamp:  start,
			Method:     r.Method,
			Path:       r.URL.Path,
			Query:      r.URL.RawQuery,
			StatusCode: rec.status,
			Message:    o.Message,
			Hint:       o.Hint,
			Success:    o.Success,
			Step:       o.Step,
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			Headers:    relevantHeaders(r),
			Body:       o.Body,
			Duration:   time.Since(start),
			Bytes:      rec.bytes,
		})
	})
}

func relevantHeaders(r *http.Request) map[string]string {
	out := map[string]string{}
	for _, name := range []string{"X-Transmission", "Content-Type", "Accept"} {
		if v := r.Header.Get(name); v != "" {
			out[name] = v
		}
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		out["Authorization"] = maskAuthorization(auth)
	}
	return out
}

// maskAuthorization exibe o usuário do Basic Auth, mas mascara a senha.
func maskAuthorization(v string) string {
	scheme, cred, ok := strings.Cut(v, " ")
	if !ok || !strings.EqualFold(scheme, "Basic") {
		return scheme + " ****"
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cred))
	if err != nil {
		return "Basic (base64 inválido)"
	}
	user, pass, _ := strings.Cut(string(raw), ":")
	return fmt.Sprintf("Basic %s:%s", user, strings.Repeat("*", len(pass)))
}

// reply escreve JSON indentado (legível no cURL) e registra o diagnóstico.
func (s *ServerChallenge2) reply(w http.ResponseWriter, r *http.Request, status int, payload any, res outcome) {
	o := outcomeFrom(r)
	body := o.Body
	*o = res
	if o.Body == "" {
		o.Body = body
	}
	if o.Step > 0 {
		s.updateStep(o.Step)
	}

	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Signal-Step", fmt.Sprintf("%d/%d", s.GetStepReached(), TotalSteps))
	if res.Hint != "" {
		h.Set("X-Hint", asciiHeader(res.Hint))
	}
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

// asciiHeader remove acentos/caracteres não-ASCII para valores de cabeçalho.
func asciiHeader(s string) string {
	repl := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i",
		"ó", "o", "ô", "o", "õ", "o", "ú", "u", "ç", "c",
		"Á", "A", "À", "A", "Â", "A", "Ã", "A", "É", "E", "Ê", "E", "Í", "I",
		"Ó", "O", "Ô", "O", "Õ", "O", "Ú", "U", "Ç", "C",
	)
	s = repl.Replace(s)
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= 0x20 && c < 0x7f {
			b = append(b, c)
		}
	}
	return string(b)
}

func (s *ServerChallenge2) methodNotAllowed(w http.ResponseWriter, r *http.Request, allowed, hint string) {
	w.Header().Set("Allow", allowed)
	s.reply(w, r, http.StatusMethodNotAllowed, map[string]string{
		"error":    fmt.Sprintf("MÉTODO %s NÃO SUPORTADO NESTA FREQUÊNCIA (405)", r.Method),
		"esperado": allowed,
		"dica":     hint,
	}, outcome{
		Message: fmt.Sprintf("Método %s recusado (esperado %s)", r.Method, allowed),
		Hint:    hint,
	})
}

// ---------------------------------------------------------------------------
// Rotas
// ---------------------------------------------------------------------------

func (s *ServerChallenge2) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/weather-service", s.handleWeatherService)
	mux.HandleFunc("/skywatching", s.handleSkywatching)
	mux.HandleFunc("/rejoice", s.handleRejoice)
	return s.instrument(mux)
}

// 1. Rota raiz da emissora
func (s *ServerChallenge2) handleRoot(w http.ResponseWriter, r *http.Request) {
	base := s.URL()
	if r.URL.Path != "/" {
		hint := "Sintonize a raiz da emissora: curl " + base + "/"
		s.reply(w, r, http.StatusNotFound, map[string]string{
			"error": "sinal analógico perdido (estática pura)",
			"hint":  hint,
		}, outcome{Message: "Estática pura: rota inexistente", Hint: hint})
		return
	}

	hint := "curl -i " + base + "/weather-service"
	s.reply(w, r, http.StatusOK, map[string]any{
		"station":        "LOCAL 58 (WCLV-TV)",
		"carrier":        "ANALOG UHF 476.000 MHz",
		"eas_alert":      "THE METEOROLOGICAL EVENT IS IN EFFECT",
		"instruction":    "DO NOT LOOK AT THE NIGHT SKY. AVOID ALL MIRRORS.",
		"transmission":   "INTERRUPTED",
		"emergency_feed": "GET /weather-service",
		"hint":           hint,
		"tip":            "Use 'curl -i' para ver o código HTTP e os cabeçalhos X-Hint / X-Signal-Step.",
	}, outcome{Message: "Sinal de abertura interceptado: AVISO METEOROLÓGICO", Hint: hint, Success: true})
}

// 2. Redirecionamento 302 em /weather-service -> /skywatching
func (s *ServerChallenge2) handleWeatherService(w http.ResponseWriter, r *http.Request) {
	hint := "Siga o desvio com -L: curl -L " + s.URL() + "/weather-service"
	w.Header().Set("Location", "/skywatching")
	s.reply(w, r, http.StatusFound, map[string]string{
		"status":   "SINAL SEQUESTRADO PELA PORTADORA LUNAR (302)",
		"redirect": "/skywatching",
		"message":  "A transmissão meteorológica foi forçada para a visualização do céu noturno.",
		"hint":     hint,
	}, outcome{
		Message: "Sinal desviado para /skywatching (HTTP 302)",
		Hint:    hint,
		Success: true,
		Step:    StepRedirect,
	})
}

// 3. Sintonia com Query Param (?sky=visible) e Header (X-Transmission: 476MHz)
func (s *ServerChallenge2) handleSkywatching(w http.ResponseWriter, r *http.Request) {
	base := s.URL()
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		s.methodNotAllowed(w, r, http.MethodGet, "Esta rota só escuta GET: remova -X/-d da chamada.")
		return
	}

	rawSky := r.URL.Query().Get("sky")
	rawTx := r.Header.Get("X-Transmission")
	sky := strings.ToLower(strings.TrimSpace(rawSky))
	transmission := strings.ToLower(strings.TrimSpace(rawTx))
	received := map[string]string{"?sky": orNone(rawSky), "X-Transmission": orNone(rawTx)}

	if sky != "visible" {
		hint := "Aponte a lente para o céu: adicione ?sky=visible na URL (use aspas na URL)."
		s.reply(w, r, http.StatusBadRequest, map[string]any{
			"error":    "VISUALIZAÇÃO DO CÉU NOTURNO NÃO ATIVADA",
			"detalhe":  "Aponte a lente para o céu adicionando ?sky=visible na URL.",
			"recebido": received,
			"exemplo":  "curl '" + base + "/skywatching?sky=visible'",
		}, outcome{Message: "Céu não visível: query ?sky=visible ausente ou incorreta", Hint: hint})
		return
	}

	if transmission != "476mhz" {
		hint := "Calibre a portadora com o cabeçalho -H 'X-Transmission: 476MHz'."
		s.reply(w, r, http.StatusBadRequest, map[string]any{
			"error":    "PORTADORA ANALÓGICA DESCALIBRADA",
			"detalhe":  "Sintonize a frequência UHF de 476MHz adicionando o cabeçalho 'X-Transmission: 476MHz'.",
			"recebido": received,
			"exemplo":  "curl -H 'X-Transmission: 476MHz' '" + base + "/skywatching?sky=visible'",
		}, outcome{Message: "Frequência descalibrada: cabeçalho X-Transmission ausente/incorreto", Hint: hint})
		return
	}

	hint := "Identifique-se em /rejoice com Basic Auth (-u viewer:moon) e envie POST JSON."
	s.reply(w, r, http.StatusOK, map[string]any{
		"transmission": "WCLV SKYWATCHING (NOVEMBER 1999)",
		"focus":        "HIS THRONE",
		"lunar_view":   "CLOAKED IN NIGHT. IT IS MOVING. IT IS AWAKE.",
		"whisper":      "IF YOU ARE AFRAID WE WILL LOOK TOGETHER.",
		"next_step":    "Autentique-se como espectador em /rejoice usando Basic Auth com usuário 'viewer' e senha 'moon' (-u viewer:moon), enviando POST com o corpo JSON: {\"action\": \"LOOK_AT_THE_MOON\"}",
		"exemplo":      "curl -u viewer:moon -X POST -H 'Content-Type: application/json' -d '{\"action\": \"LOOK_AT_THE_MOON\"}' " + base + "/rejoice",
	}, outcome{
		Message: "HIS THRONE: foco telescópico fixado na Lua",
		Hint:    hint,
		Success: true,
		Step:    StepTuned,
	})
}

// 4. Aceitação da Lua via POST (/rejoice)
func (s *ServerChallenge2) handleRejoice(w http.ResponseWriter, r *http.Request) {
	base := s.URL()
	if r.Method != http.MethodPost {
		s.methodNotAllowed(w, r, http.MethodPost,
			"Envie POST: use -X POST (ou apenas -d '...', que já implica POST).")
		return
	}

	user, pass, ok := r.BasicAuth()
	if !ok || user != "viewer" || pass != "moon" {
		msg := "Tentativa sem credenciais de espectador"
		if ok {
			msg = fmt.Sprintf("Credenciais recusadas para o usuário '%s'", user)
		}
		hint := "Autentique-se com Basic Auth: curl -u viewer:moon ..."
		w.Header().Set("WWW-Authenticate", `Basic realm="Local 58 Moon Feed"`)
		s.reply(w, r, http.StatusUnauthorized, map[string]string{
			"error":   "ACESSO NEGADO (401)",
			"detalhe": "Você precisa se autenticar como espectador com -u viewer:moon.",
			"exemplo": "curl -u viewer:moon -X POST ... " + base + "/rejoice",
		}, outcome{Message: msg, Hint: hint})
		return
	}

	bodyBytes, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	preview := previewBody(bodyBytes)
	outcomeFrom(r).Body = preview
	if err != nil {
		s.reply(w, r, http.StatusRequestEntityTooLarge, map[string]string{
			"error": "SINAL GRANDE DEMAIS (limite de 64KB)",
		}, outcome{Message: "Corpo da requisição excedeu o limite", Step: StepAuth, Success: true})
		return
	}

	example := "curl -u viewer:moon -H 'Content-Type: application/json' -d '{\"action\": \"LOOK_AT_THE_MOON\"}' " + base + "/rejoice"

	var payload struct {
		Action string `json:"action"`
	}
	trimmed := strings.TrimSpace(string(bodyBytes))
	switch {
	case trimmed == "":
		hint := "Corpo vazio: envie o JSON com -d '{\"action\": \"LOOK_AT_THE_MOON\"}'."
		s.reply(w, r, http.StatusBadRequest, map[string]string{
			"error":    "NENHUM SINAL NO CORPO DA REQUISIÇÃO",
			"esperado": `{"action": "LOOK_AT_THE_MOON"}`,
			"exemplo":  example,
		}, outcome{Message: "Espectador identificado, mas o corpo veio vazio", Hint: hint, Success: true, Step: StepAuth})
		return
	case json.Unmarshal(bodyBytes, &payload) != nil:
		hint := "JSON malformado: envolva o corpo em aspas simples e use aspas duplas nas chaves."
		s.reply(w, r, http.StatusBadRequest, map[string]string{
			"error":    "SINAL CORROMPIDO (JSON INVÁLIDO)",
			"esperado": `{"action": "LOOK_AT_THE_MOON"}`,
			"recebido": preview,
			"exemplo":  example,
		}, outcome{Message: "Espectador identificado, mas o JSON está malformado", Hint: hint, Success: true, Step: StepAuth})
		return
	case strings.ToUpper(strings.TrimSpace(payload.Action)) != "LOOK_AT_THE_MOON":
		hint := "O campo \"action\" deve ser exatamente LOOK_AT_THE_MOON."
		s.reply(w, r, http.StatusBadRequest, map[string]string{
			"error":    "COMANDO INVÁLIDO",
			"esperado": `{"action": "LOOK_AT_THE_MOON"}`,
			"recebido": preview,
		}, outcome{Message: fmt.Sprintf("Espectador identificado, ação '%s' recusada", payload.Action), Hint: hint, Success: true, Step: StepAuth})
		return
	}

	s.updateStep(StepAuth)
	s.reply(w, r, http.StatusOK, map[string]any{
		"status":       "REJOICE",
		"transmission": "VOCÊ OLHOU. A LUZ AGORA HABITA EM VOCÊ.",
		"flag":         Flag2,
	}, outcome{
		Message: "REJOICE! VOCÊ OLHOU PARA A LUA! FLAG 2 LIBERADA",
		Hint:    "Copie a flag e decodifique no console (tecla f).",
		Success: true,
		Step:    StepRejoice,
	})
}

func orNone(v string) string {
	if v == "" {
		return "(ausente)"
	}
	return v
}

func previewBody(b []byte) string {
	s := strings.Join(strings.Fields(string(b)), " ")
	if len(s) > 120 {
		return s[:117] + "..."
	}
	return s
}

// ---------------------------------------------------------------------------
// Ciclo de vida
// ---------------------------------------------------------------------------

func (s *ServerChallenge2) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		return nil
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		return fmt.Errorf("falha ao iniciar servidor HTTP na porta %d: %w", s.port, err)
	}

	s.listener = listener
	s.server = &http.Server{
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	s.isRunning = true

	srv := s.server
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.mu.Lock()
			s.isRunning = false
			s.mu.Unlock()
		}
	}()

	return nil
}

func (s *ServerChallenge2) updateStep(step int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if step > s.StepReached {
		s.StepReached = step
	}
}

func (s *ServerChallenge2) GetStepReached() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.StepReached
}

func (s *ServerChallenge2) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isRunning || s.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
	s.isRunning = false
}

// emit publica o evento sem nunca bloquear o handler HTTP: se o buffer estiver
// cheio (TUI lenta), descarta o evento mais antigo.
func (s *ServerChallenge2) emit(evt HTTPLogEvent) {
	for i := 0; i < 3; i++ {
		select {
		case s.Events <- evt:
			return
		default:
			select {
			case <-s.Events:
			default:
			}
		}
	}
}
