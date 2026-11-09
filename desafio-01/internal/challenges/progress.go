package challenges

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Flags autênticas do universo de Local 58 (Kris Straub)
const (
	Flag1 = "FLAG{IF_YOU_ARE_AFRAID_WE_WILL_LOOK_TOGETHER}"
	Flag2 = "FLAG{HIS_THRONE_REJOICE_LOOK_AT_THE_MOON}"
	Flag3 = "FLAG{THERE_ARE_NO_FACES_THERE_ARE_NO_FACES}"
	Flag4 = "FLAG{HOC_VOLUNTAS_TUA_PROCEDURE_3F}"
)

// TotalChallenges é a quantidade de transmissões (desafios) do console.
const TotalChallenges = 4

var flagByChallenge = map[int]string{1: Flag1, 2: Flag2, 3: Flag3, 4: Flag4}

// SolvedQuotes são as frases de Local 58 exibidas quando cada sinal é decodificado.
var SolvedQuotes = map[int]string{
	1: "SE VOCÊ ESTIVER COM MEDO, NÓS OLHAREMOS JUNTOS.",
	2: "HIS THRONE. REJOICE.",
	3: "THERE ARE NO FACES. THERE ARE NO FACES.",
	4: "HOC VOLUNTAS TUA. PROCEDIMENTO 3F.",
}

// SubmitKind classifica o resultado de uma tentativa de decodificação,
// permitindo que a interface explique o PORQUÊ de uma recusa.
type SubmitKind int

const (
	SubmitCorrect   SubmitKind = iota // flag válida e inédita
	SubmitAlready                     // flag válida, já registrada antes
	SubmitEmpty                       // nada foi digitado
	SubmitBadFormat                   // não segue o formato FLAG{...}
	SubmitDecoy                       // flag falsa plantada no acervo (FAKE_FLAG{...})
	SubmitWrongCase                   // conteúdo certo, maiúsculas/minúsculas erradas
	SubmitWrong                       // formato correto, mas não reconhecida
)

// SubmitResult descreve o resultado de uma submissão de flag.
type SubmitResult struct {
	Kind        SubmitKind
	ChallengeID int    // 0 quando nenhuma transmissão foi reconhecida
	Flag        string // flag normalizada/extraída da entrada
	Message     string
}

// OK indica se a flag é válida (inédita ou já registrada).
func (r SubmitResult) OK() bool { return r.Kind == SubmitCorrect || r.Kind == SubmitAlready }

var (
	flagPattern  = regexp.MustCompile(`\bFLAG\{[^{}\s]*\}`)
	decoyPattern = regexp.MustCompile(`(?i)FAKE_FLAG\{[^{}\s]*\}`)
	loosePattern = regexp.MustCompile(`(?i)\bflag\{[^{}\s]*\}`)
)

type Progress struct {
	mu            sync.Mutex
	Completed     map[int]bool      `json:"completed"`
	FlagsFound    map[int]string    `json:"flags_found"`
	HintsRevealed map[int]int       `json:"hints_revealed"`
	StartedAt     time.Time         `json:"started_at"`
	CompletedAt   map[int]time.Time `json:"completed_at"`
	filePath      string
}

func NewProgress(dataDir string) *Progress {
	p := &Progress{
		Completed:     map[int]bool{1: false, 2: false, 3: false, 4: false},
		FlagsFound:    make(map[int]string),
		HintsRevealed: make(map[int]int),
		CompletedAt:   make(map[int]time.Time),
		filePath:      filepath.Join(dataDir, ".progress.json"),
	}
	p.load()
	if p.StartedAt.IsZero() {
		p.mu.Lock()
		p.StartedAt = time.Now()
		_ = p.save()
		p.mu.Unlock()
	}
	return p
}

func (p *Progress) load() {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return
	}

	var saved struct {
		Completed     map[int]bool      `json:"completed"`
		FlagsFound    map[int]string    `json:"flags_found"`
		HintsRevealed map[int]int       `json:"hints_revealed"`
		StartedAt     time.Time         `json:"started_at"`
		CompletedAt   map[int]time.Time `json:"completed_at"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}
	for k, v := range saved.Completed {
		p.Completed[k] = v
	}
	for k, v := range saved.FlagsFound {
		p.FlagsFound[k] = v
	}
	for k, v := range saved.HintsRevealed {
		p.HintsRevealed[k] = v
	}
	for k, v := range saved.CompletedAt {
		p.CompletedAt[k] = v
	}
	p.StartedAt = saved.StartedAt
}

func (p *Progress) save() error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.filePath, data, 0644)
}

// normalizeFlagInput limpa a entrada colada pelo jogador. Aceita colagens
// "sujas" como `"flag": "FLAG{...}",` ou `TOWER_OVERRIDE_KEY=FLAG{...}` e
// extrai somente o trecho FLAG{...}.
func normalizeFlagInput(input string) string {
	trimmed := strings.TrimSpace(input)
	if m := flagPattern.FindString(trimmed); m != "" {
		return m
	}
	return strings.Trim(trimmed, "\"'` ,;")
}

// Submit valida uma flag e retorna um resultado detalhado.
func (p *Progress) Submit(input string) SubmitResult {
	p.mu.Lock()
	defer p.mu.Unlock()

	raw := strings.TrimSpace(input)
	if raw == "" {
		return SubmitResult{Kind: SubmitEmpty, Message: "Nenhum sinal recebido. Cole ou digite a chave FLAG{...}."}
	}

	if decoyPattern.MatchString(raw) && !flagPattern.MatchString(raw) {
		return SubmitResult{
			Kind:    SubmitDecoy,
			Message: "SINAL FALSO. Esta chave foi plantada no acervo para enganar operadores. A Lua está rindo de você.",
		}
	}

	flag := normalizeFlagInput(raw)

	for id, real := range flagByChallenge {
		if flag != real {
			continue
		}
		if p.Completed[id] {
			return SubmitResult{Kind: SubmitAlready, ChallengeID: id, Flag: flag,
				Message: "Este sinal já foi decodificado anteriormente."}
		}
		p.Completed[id] = true
		p.FlagsFound[id] = flag
		p.CompletedAt[id] = time.Now()
		_ = p.save()
		return SubmitResult{Kind: SubmitCorrect, ChallengeID: id, Flag: flag,
			Message: "SINAL " + string(rune('0'+id)) + " CAPTADO: '" + SolvedQuotes[id] + "'"}
	}

	if loose := loosePattern.FindString(raw); loose != "" {
		for id, real := range flagByChallenge {
			if strings.EqualFold(loose, real) {
				return SubmitResult{Kind: SubmitWrongCase, ChallengeID: id, Flag: loose,
					Message: "Quase... a portadora reconhece o padrão, mas as chaves são sensíveis a maiúsculas/minúsculas."}
			}
		}
		if !flagPattern.MatchString(raw) {
			return SubmitResult{Kind: SubmitBadFormat, Flag: loose,
				Message: "Formato inválido: o prefixo deve ser exatamente 'FLAG' em maiúsculas, como FLAG{...}."}
		}
	}

	if !flagPattern.MatchString(flag) {
		return SubmitResult{Kind: SubmitBadFormat, Flag: flag,
			Message: "Formato inválido. Toda chave interceptada segue o padrão FLAG{...} (sem espaços)."}
	}

	return SubmitResult{Kind: SubmitWrong, Flag: flag,
		Message: "Sinal não reconhecido. A transmissão não reconhece este código."}
}

// SubmitFlag mantém a assinatura legada: (id, válida, jáRegistrada, mensagem).
func (p *Progress) SubmitFlag(input string) (int, bool, bool, string) {
	r := p.Submit(input)
	return r.ChallengeID, r.OK(), r.Kind == SubmitAlready, r.Message
}

// IsCompleted informa se a transmissão foi decodificada (thread-safe).
func (p *Progress) IsCompleted(id int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Completed[id]
}

// HintsUsed retorna quantas dicas da transmissão já foram reveladas.
func (p *Progress) HintsUsed(id int) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.HintsRevealed[id]
}

// TotalHintsUsed soma as dicas reveladas em todas as transmissões.
func (p *Progress) TotalHintsUsed() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	total := 0
	for _, n := range p.HintsRevealed {
		total += n
	}
	return total
}

// RevealHint revela a próxima dica (até max) e persiste. Retorna o total revelado.
func (p *Progress) RevealHint(id, max int) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.HintsRevealed[id] < max {
		p.HintsRevealed[id]++
		_ = p.save()
	}
	return p.HintsRevealed[id]
}

// Elapsed retorna o tempo entre o início da investigação e a última decodificação
// (ou agora, se ainda houver sinais pendentes).
func (p *Progress) Elapsed(now time.Time) time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	end := now
	if p.Completed[1] && p.Completed[2] && p.Completed[3] && p.Completed[4] {
		end = time.Time{}
		for _, t := range p.CompletedAt {
			if t.After(end) {
				end = t
			}
		}
	}
	if p.StartedAt.IsZero() || end.Before(p.StartedAt) {
		return 0
	}
	return end.Sub(p.StartedAt)
}

func (p *Progress) IsAllCompleted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Completed[1] && p.Completed[2] && p.Completed[3] && p.Completed[4]
}

func (p *Progress) TotalCompleted() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	count := 0
	for i := 1; i <= TotalChallenges; i++ {
		if p.Completed[i] {
			count++
		}
	}
	return count
}
