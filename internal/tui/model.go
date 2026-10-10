package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/devops/desafio-analog-horror/internal/challenges"
)

type Screen int

const (
	ScreenOverview Screen = iota // CH 00 — central
	ScreenChallenge1
	ScreenChallenge2
	ScreenChallenge3
	ScreenChallenge4
	ScreenSubmitFlag // CH 05 — decodificador
	ScreenVictory
)

// Aliases semânticos usados pela interface.
const (
	ScreenCentral = ScreenOverview
	ScreenDecoder = ScreenSubmitFlag
	channelCount  = 6 // CH 00..05
)

const (
	frameInterval = 150 * time.Millisecond
	tuningFrames  = 3
	toastDuration = 6 * time.Second
	maxLogs       = 200
	maxAttempts   = 6
)

type toastKind int

const (
	toastInfo toastKind = iota
	toastSuccess
	toastWarn
	toastError
)

type toast struct {
	text  string
	kind  toastKind
	until time.Time
}

type attempt struct {
	at     time.Time
	input  string
	result challenges.SubmitResult
}

// Config agrupa as dependências do console.
type Config struct {
	Progress     *challenges.Progress
	Server       *challenges.ServerChallenge2
	WorkspaceDir string
	ServerErr    error // falha ao subir o transmissor HTTP (ex.: porta ocupada)
	NoGlitch     bool  // desliga animações/estática (acessibilidade)
}

type envState struct {
	archiveOK bool
	repoOK    bool
}

type Model struct {
	cfg      Config
	Progress *challenges.Progress
	Server2  *challenges.ServerChallenge2

	catalog []Challenge
	keys    keyMap
	help    help.Model

	screen     Screen
	prevScreen Screen
	showHelp   bool

	width, height int
	ready         bool
	frame         int
	now           time.Time
	tuning        int

	viewport viewport.Model
	input    textinput.Model

	logs       []challenges.HTTPLogEvent
	serverStep int
	attempts   []attempt
	lastResult *challenges.SubmitResult
	toast      *toast

	env   envState
	image challenges.ImageStatus
}

// NewModel cria o console a partir da configuração.
func NewModel(cfg Config) Model {
	ti := textinput.New()
	ti.Placeholder = "FLAG{...}"
	ti.Prompt = "▶ "
	ti.PromptStyle = AccentStyle
	ti.TextStyle = TextStyle
	ti.PlaceholderStyle = DimStyle
	ti.Cursor.Style = GreenStyle
	ti.CharLimit = 256
	ti.Width = 50

	vp := viewport.New(80, 20)
	vp.KeyMap = viewport.KeyMap{} // rolagem tratada manualmente (evita conflito com h/f/d/b)

	h := help.New()
	h.ShortSeparator = DimStyle.Render(" · ")
	h.Styles.ShortKey = LabelStyle
	h.Styles.ShortDesc = MutedStyle
	h.Styles.FullKey = LabelStyle
	h.Styles.FullDesc = MutedStyle
	h.Styles.ShortSeparator = DimStyle
	h.Styles.FullSeparator = DimStyle

	m := Model{
		cfg:      cfg,
		Progress: cfg.Progress,
		Server2:  cfg.Server,
		catalog:  buildCatalog(cfg.WorkspaceDir, cfg.Server.URL()),
		keys:     newKeyMap(),
		help:     h,
		screen:   ScreenCentral,
		width:    100,
		height:   30,
		now:      time.Now(),
		viewport: vp,
		input:    ti,
		logs:     make([]challenges.HTTPLogEvent, 0, maxLogs),
		image:    challenges.ImageStatus{State: challenges.ImageChecking, Detail: "verificando imagem..."},
	}
	m.env = scanEnv(cfg.WorkspaceDir)
	if cfg.ServerErr != nil {
		m.pushToast("Transmissor HTTP offline: "+cfg.ServerErr.Error(), toastError)
	}
	if cfg.Progress.IsAllCompleted() {
		m.screen = ScreenVictory
	}
	m.syncLayout()
	return m
}

func scanEnv(workspace string) envState {
	exists := func(p string) bool { _, err := os.Stat(p); return err == nil }
	return envState{
		archiveOK: exists(filepath.Join(workspace, "wclv_archive")),
		repoOK:    exists(filepath.Join(workspace, "wclv_station_repo", ".git")),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tickCmd(),
		waitForHTTPEvent(m.Server2.Events),
		checkImageCmd(m.cfg.WorkspaceDir),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(frameInterval, func(t time.Time) tea.Msg { return TickMsg(t) })
}

func waitForHTTPEvent(events chan challenges.HTTPLogEvent) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-events
		if !ok {
			return nil
		}
		return HTTPLogMsg(evt)
	}
}

func checkImageCmd(workspace string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		return ImageStatusMsg(challenges.EnsureContingencyImage(ctx, workspace))
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true

	case TickMsg:
		m.now = time.Time(msg)
		m.frame++
		if m.tuning > 0 {
			m.tuning--
		}
		if m.toast != nil && m.now.After(m.toast.until) {
			m.toast = nil
		}
		cmds = append(cmds, tickCmd())

	case HTTPLogMsg:
		m.handleHTTPEvent(challenges.HTTPLogEvent(msg))
		cmds = append(cmds, waitForHTTPEvent(m.Server2.Events))

	case ImageStatusMsg:
		m.image = challenges.ImageStatus(msg)
		switch m.image.State {
		case challenges.ImageReady:
			if strings.Contains(m.image.Detail, "construída") {
				m.pushToast("Imagem "+challenges.ContingencyImage+" construída. Transmissão 4 disponível.", toastSuccess)
			}
		case challenges.ImageChecking, challenges.ImageUnknown:
		default:
			m.pushToast("Transmissão 4 indisponível: "+m.image.Detail+" (tecla r para tentar de novo)", toastWarn)
		}

	case tea.KeyMsg:
		if cmd, handled := m.handleKey(msg); handled {
			m.syncLayout()
			return m, cmd
		}
	}

	// Mensagens não tratadas (ex.: piscar do cursor) seguem para o input.
	if m.screen == ScreenDecoder {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.syncLayout()
	return m, tea.Batch(cmds...)
}

func (m *Model) handleHTTPEvent(evt challenges.HTTPLogEvent) {
	m.logs = append(m.logs, evt)
	if len(m.logs) > maxLogs {
		m.logs = m.logs[len(m.logs)-maxLogs:]
	}
	if evt.Step > m.serverStep {
		m.serverStep = evt.Step
		labels := map[int]string{
			challenges.StepRedirect: "desvio 302 seguido",
			challenges.StepTuned:    "portadora sintonizada — foco na Lua",
			challenges.StepAuth:     "espectador identificado",
			challenges.StepRejoice:  "REJOICE! Flag liberada no seu terminal — tecle f para decodificar",
		}
		kind := toastInfo
		if evt.Step == challenges.StepRejoice {
			kind = toastSuccess
		}
		m.pushToast(fmt.Sprintf("📡 CH 02 · ESTÁGIO %d/%d: %s", evt.Step, challenges.TotalSteps, labels[evt.Step]), kind)
	}
}

// handleKey retorna (cmd, true) quando a tecla foi consumida.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	k := m.keys

	if key.Matches(msg, k.ForceQ) {
		m.Server2.Stop()
		return tea.Quit, true
	}

	if m.showHelp {
		if key.Matches(msg, k.Back, k.Help, k.Quit) {
			m.showHelp = false
		}
		return nil, true
	}

	// No decodificador, o campo de texto tem prioridade sobre letras/números.
	if m.screen == ScreenDecoder {
		switch {
		case key.Matches(msg, k.Submit):
			m.submitFlag()
			return nil, true
		case key.Matches(msg, k.Back):
			m.goBack()
			return nil, true
		case msg.String() == "tab":
			m.switchChannel(+1)
			return nil, true
		case msg.String() == "shift+tab":
			m.switchChannel(-1)
			return nil, true
		case msg.Type == tea.KeyUp || msg.Type == tea.KeyDown || msg.Type == tea.KeyPgUp || msg.Type == tea.KeyPgDown:
			m.scroll(msg)
			return nil, true
		}
		return nil, false // deixa o textinput tratar
	}

	switch {
	case key.Matches(msg, k.Quit):
		m.Server2.Stop()
		return tea.Quit, true
	case key.Matches(msg, k.Help):
		m.showHelp = true
	case key.Matches(msg, k.Back):
		m.goBack()
	case key.Matches(msg, k.NextCh):
		m.switchChannel(+1)
	case key.Matches(msg, k.PrevCh):
		m.switchChannel(-1)
	case key.Matches(msg, k.Jump):
		m.gotoScreen(Screen(msg.Runes[0] - '0'))
	case key.Matches(msg, k.Decode):
		m.gotoScreen(ScreenDecoder)
	case key.Matches(msg, k.Hint):
		m.revealHint()
	case key.Matches(msg, k.Rescan):
		m.env = scanEnv(m.cfg.WorkspaceDir)
		m.image = challenges.ImageStatus{State: challenges.ImageChecking, Detail: "verificando imagem..."}
		m.pushToast("Re-escaneando acervo, repositório e imagem Docker...", toastInfo)
		return checkImageCmd(m.cfg.WorkspaceDir), true
	case key.Matches(msg, k.Top):
		m.viewport.GotoTop()
	case key.Matches(msg, k.Bottom):
		m.viewport.GotoBottom()
	case key.Matches(msg, k.Up, k.Down, k.PageUp, k.PageDown):
		m.scroll(msg)
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) scroll(msg tea.KeyMsg) {
	k := m.keys
	switch {
	case key.Matches(msg, k.Up):
		m.viewport.ScrollUp(1)
	case key.Matches(msg, k.Down):
		m.viewport.ScrollDown(1)
	case key.Matches(msg, k.PageUp):
		m.viewport.HalfPageUp()
	case key.Matches(msg, k.PageDown):
		m.viewport.HalfPageDown()
	}
}

func (m *Model) gotoScreen(s Screen) {
	if s < 0 || s > ScreenVictory {
		return
	}
	if s == m.screen {
		return
	}
	m.prevScreen = m.screen
	m.screen = s
	m.viewport.GotoTop()
	if !m.cfg.NoGlitch {
		m.tuning = tuningFrames
	}
	if s == ScreenDecoder {
		m.input.Focus()
	} else {
		m.input.Blur()
	}
}

func (m *Model) switchChannel(delta int) {
	cur := int(m.screen)
	if m.screen == ScreenVictory {
		cur = int(ScreenCentral)
	}
	next := (cur + delta + channelCount) % channelCount
	m.gotoScreen(Screen(next))
}

func (m *Model) goBack() {
	if m.screen == ScreenCentral {
		return
	}
	target := m.prevScreen
	if target == m.screen || target == ScreenDecoder || target == ScreenVictory {
		target = ScreenCentral
	}
	m.gotoScreen(target)
}

func (m *Model) revealHint() {
	ch := m.currentChallenge()
	if ch == nil {
		m.pushToast("Sintonize uma transmissão (1-4) para revelar dicas.", toastInfo)
		return
	}
	before := m.Progress.HintsUsed(ch.ID)
	after := m.Progress.RevealHint(ch.ID, len(ch.Hints))
	if after == before {
		m.pushToast("Todas as dicas desta transmissão já foram reveladas.", toastInfo)
		return
	}
	m.pushToast(fmt.Sprintf("Dica %d/%d de CH %02d decodificada.", after, len(ch.Hints), ch.ID), toastWarn)
	m.syncLayout()
	m.viewport.GotoBottom()
}

func (m *Model) submitFlag() {
	value := m.input.Value()
	res := m.Progress.Submit(value)
	m.lastResult = &res
	if res.Kind != challenges.SubmitEmpty {
		m.attempts = append(m.attempts, attempt{at: m.now, input: strings.TrimSpace(value), result: res})
		if len(m.attempts) > maxAttempts {
			m.attempts = m.attempts[len(m.attempts)-maxAttempts:]
		}
		m.input.SetValue("")
	}

	switch res.Kind {
	case challenges.SubmitCorrect:
		m.pushToast(fmt.Sprintf("✓ SINAL %d CAPTADO — %s", res.ChallengeID, challenges.SolvedQuotes[res.ChallengeID]), toastSuccess)
		if m.Progress.IsAllCompleted() {
			m.gotoScreen(ScreenVictory)
		}
	case challenges.SubmitAlready:
		m.pushToast(res.Message, toastInfo)
	case challenges.SubmitWrongCase, challenges.SubmitBadFormat:
		m.pushToast(res.Message, toastWarn)
	default:
		m.pushToast(res.Message, toastError)
	}
}

func (m *Model) pushToast(text string, kind toastKind) {
	m.toast = &toast{text: text, kind: kind, until: m.now.Add(toastDuration)}
}

func (m Model) currentChallenge() *Challenge {
	if m.screen >= ScreenChallenge1 && m.screen <= ScreenChallenge4 {
		return &m.catalog[int(m.screen)-1]
	}
	return nil
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

type layout struct {
	headerH, barsH, footerH int
	bodyH                   int
	sidebarW, mainW         int
	logH                    int // altura do painel de log (CH 02)
	panelH                  int // altura do painel principal
}

func (m Model) computeLayout() layout {
	l := layout{headerH: 1, footerH: 2}
	if m.height >= 24 && !m.cfg.NoGlitch {
		l.barsH = 1
	}
	l.bodyH = max(m.height-l.headerH-l.barsH-l.footerH, 3)
	if m.width >= 100 {
		l.sidebarW = 30
	}
	l.mainW = m.width - l.sidebarW
	l.panelH = l.bodyH
	if m.screen == ScreenChallenge2 && !m.showHelp {
		switch {
		case l.bodyH >= 30:
			l.logH = 12
		case l.bodyH >= 20:
			l.logH = 8
		}
		l.panelH = l.bodyH - l.logH
	}
	return l
}

// syncLayout redimensiona o viewport e recalcula o conteúdo da tela atual.
// Chamado ao fim de todo Update, mantendo View() livre de efeitos colaterais.
func (m *Model) syncLayout() {
	l := m.computeLayout()
	w := contentWidth(PanelStyle, l.mainW)
	h := max(l.panelH-PanelStyle.GetVerticalFrameSize()-1, 1) // -1: linha de título do painel
	m.viewport.Width = w
	m.viewport.Height = h
	m.input.Width = max(min(w-8, 90), 10)

	wasBottom := m.viewport.AtBottom() && m.viewport.YOffset > 0
	m.viewport.SetContent(m.renderContent(w))
	if wasBottom {
		m.viewport.GotoBottom()
	}
}

// renderContent produz o corpo rolável da tela atual.
func (m Model) renderContent(w int) string {
	if m.showHelp {
		return m.viewHelp(w)
	}
	switch m.screen {
	case ScreenChallenge1, ScreenChallenge2, ScreenChallenge3, ScreenChallenge4:
		return m.viewChallenge(*m.currentChallenge(), w)
	case ScreenDecoder:
		return m.viewDecoder(w)
	case ScreenVictory:
		return m.viewVictory(w)
	default:
		return m.viewCentral(w)
	}
}

// statusTag mantém a etiqueta de status reutilizada em várias telas.
func (m Model) statusTag(challengeID int) string {
	if m.Progress.IsCompleted(challengeID) {
		return TagCompletedStyle.Render("✓ DECODIFICADO")
	}
	return TagPendingStyle.Render("◌ PENDENTE")
}

func (m Model) View() string {
	if !m.ready {
		return "\n  " + AccentStyle.Render("SINTONIZANDO UHF 476.000 MHz...")
	}
	if m.width < 60 || m.height < 16 {
		msg := fmt.Sprintf("SINAL FRACO\n\nO monitor precisa de pelo menos 60x16.\nAtual: %dx%d", m.width, m.height)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, AlertBoxStyle.Render(msg))
	}

	l := m.computeLayout()
	parts := []string{m.renderHeader()}
	if l.barsH > 0 {
		parts = append(parts, colorBars(m.width, m.glitchActive()))
	}

	main := m.renderMainPanel(l)
	if l.logH > 0 {
		main = lipgloss.JoinVertical(lipgloss.Left, main, m.renderLogPanel(l.mainW, l.logH))
	}
	body := main
	if l.sidebarW > 0 {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.renderSidebar(l.sidebarW, l.bodyH), main)
	}
	parts = append(parts, body, m.renderFooter())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// glitchActive sorteia frames de interferência (desativado com --no-glitch).
func (m Model) glitchActive() bool {
	if m.cfg.NoGlitch {
		return false
	}
	return m.tuning > 0 || m.frame%47 == 0 || m.frame%47 == 1
}
