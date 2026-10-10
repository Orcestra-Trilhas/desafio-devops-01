package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/devops/desafio-analog-horror/internal/challenges"
)

// HeaderText é o aviso EAS exibido na central.
const HeaderText = `LOCAL 58 • WCLV-TV • CANAL 58 UHF • MASON COUNTY, WV
LIDERANÇA COMUNITÁRIA DESDE 1934
AVISO: O EVENTO METEOROLÓGICO CONTINUA ATÉ O AMANHECER.
EVITEM TODOS OS ESPELHOS. NÃO OLHEM PARA O CÉU NOTURNO.`

const crawlText = "  ⚠ O EVENTO METEOROLÓGICO CONTINUA ATÉ O AMANHECER  •  EVITEM TODOS OS ESPELHOS  •  " +
	"NÃO OLHEM PARA O CÉU NOTURNO  •  SE VOCÊ ESTIVER COM MEDO, NÓS OLHAREMOS JUNTOS  •  " +
	"WCLV-TV CANAL 58 — LIDERANÇA COMUNITÁRIA DESDE 1934  •"

// channelName devolve o rótulo curto de cada canal.
func (m Model) channelName(s Screen) string {
	switch s {
	case ScreenCentral:
		return "CENTRAL"
	case ScreenDecoder:
		return "DECODIFICAR"
	case ScreenVictory:
		return "FIM DA TRANSMISSÃO"
	default:
		return m.catalog[int(s)-1].Short
	}
}

func (m Model) channelTitle() string {
	if m.showHelp {
		return "MANUAL DO OPERADOR"
	}
	switch m.screen {
	case ScreenCentral:
		return "CH 00 · CENTRAL DE OPERAÇÕES"
	case ScreenDecoder:
		return "CH 05 · DECODIFICADOR DE SINAIS"
	case ScreenVictory:
		return "TRANSMISSÃO FINALIZADA"
	default:
		ch := m.currentChallenge()
		return fmt.Sprintf("CH %02d · %s", ch.ID, ch.Title)
	}
}

// channelStep devolve o progresso (0..4) do CH 02, considerando a flag submetida.
func (m Model) channelStep() int {
	if m.Progress.IsCompleted(2) {
		return challenges.TotalSteps
	}
	return m.Server2.GetStepReached()
}

// ---------------------------------------------------------------------------
// Cabeçalho
// ---------------------------------------------------------------------------

func (m Model) renderHeader() string {
	onAir := OnAirStyle.Render("● NO AR")
	if !m.cfg.NoGlitch && (m.frame/4)%2 == 1 {
		onAir = OffAirStyle.Render("○ NO AR")
	}
	station := AccentStyle.Render(" LOCAL 58 ") + DimStyle.Render("·") + GreenStyle.Render(" WCLV-TV ") +
		DimStyle.Render("·") + MutedStyle.Render(" UHF 476.000 MHz")

	done := m.Progress.TotalCompleted()
	srv := GreenStyle.Render(fmt.Sprintf("● :%d", m.Server2.Port()))
	if !m.Server2.Running() {
		srv = LogErrorStyle.Render(fmt.Sprintf("✗ :%d", m.Server2.Port()))
	}
	right := MutedStyle.Render("SINAIS ") + segmentBar(done, challenges.TotalChallenges) +
		TextStyle.Render(fmt.Sprintf(" %d/%d", done, challenges.TotalChallenges)) +
		DimStyle.Render("  │  ") + MutedStyle.Render("HTTP ") + srv +
		DimStyle.Render("  │  ") + LabelStyle.Render(m.now.Format("15:04:05")) + " "

	left := onAir + station
	if gap := m.width - lipgloss.Width(left) - lipgloss.Width(right); gap >= 1 {
		return left + strings.Repeat(" ", gap) + right
	}
	// Terminal estreito: prioriza status.
	left = onAir
	if gap := m.width - lipgloss.Width(left) - lipgloss.Width(right); gap >= 1 {
		return left + strings.Repeat(" ", gap) + right
	}
	return truncate(onAir+" "+right, m.width)
}

// ---------------------------------------------------------------------------
// Barra lateral (guia de canais + telemetria)
// ---------------------------------------------------------------------------

func (m Model) renderSidebar(w, h int) string {
	style := PanelStyle.BorderForeground(ColorGreenDim)
	cw := contentWidth(style, w)
	var lines []string
	rule := DimStyle.Render(strings.Repeat("─", cw))

	lines = append(lines, LabelStyle.Render("GUIA DE CANAIS"), rule)
	for s := ScreenCentral; s <= ScreenDecoder; s++ {
		lines = append(lines, m.sidebarEntry(s, cw))
	}

	done := m.Progress.TotalCompleted()
	lines = append(lines, rule,
		MutedStyle.Render("SINAIS  ")+segmentBar(done, challenges.TotalChallenges)+TextStyle.Render(fmt.Sprintf(" %d/%d", done, challenges.TotalChallenges)),
		MutedStyle.Render("DICAS   ")+TextStyle.Render(fmt.Sprintf("%d", m.Progress.TotalHintsUsed())),
		MutedStyle.Render("NO AR   ")+TextStyle.Render(formatDuration(m.Progress.Elapsed(m.now))),
		rule,
		LabelStyle.Render("TELEMETRIA"),
		m.telemetryLine("HTTP", m.httpState()),
		m.telemetryLine("DOCKER", m.imageState()),
		m.telemetryLine("ACERVO", boolState(m.env.archiveOK, "pronto", "ausente")),
		m.telemetryLine("GIT", boolState(m.env.repoOK, "pronto", "ausente")),
	)

	inner := h - style.GetVerticalFrameSize()
	if len(lines) > inner {
		lines = lines[:max(inner, 0)]
	}
	for i := range lines {
		lines[i] = truncate(lines[i], cw)
	}
	return style.Width(borderedWidth(style, w)).Height(inner).Render(strings.Join(lines, "\n"))
}

func (m Model) sidebarEntry(s Screen, cw int) string {
	name := m.channelName(s)
	skill, icon := "", ""
	switch {
	case s >= ScreenChallenge1 && s <= ScreenChallenge4:
		ch := m.catalog[int(s)-1]
		skill = ch.Skill
		switch {
		case m.Progress.IsCompleted(ch.ID):
			icon = GreenStyle.Render("✓")
		case ch.ID == 2 && m.channelStep() > 0:
			icon = LabelStyle.Render("◐")
		default:
			icon = DimStyle.Render("·")
		}
	case s == ScreenDecoder:
		skill = "[f]"
	}

	selected := s == m.screen || (m.screen == ScreenVictory && s == ScreenCentral && false)
	text := fmt.Sprintf(" %02d %-9s %-9s", int(s), truncate(name, 9), skill)
	if selected {
		plain := padRight(fmt.Sprintf("▸%s", text[1:]), cw-2)
		return MenuSelectedStyle.Render(plain) + " " + icon
	}
	return padRight(MenuItemStyle.Render(text), cw-2) + " " + icon
}

type telemetry struct {
	ok    int // 1 ok, 0 em andamento, -1 erro
	label string
}

func boolState(ok bool, yes, no string) telemetry {
	if ok {
		return telemetry{1, yes}
	}
	return telemetry{-1, no}
}

func (m Model) httpState() telemetry {
	if m.Server2.Running() {
		return telemetry{1, fmt.Sprintf("online :%d", m.Server2.Port())}
	}
	return telemetry{-1, "offline"}
}

func (m Model) imageState() telemetry {
	switch m.image.State {
	case challenges.ImageReady:
		return telemetry{1, "imagem pronta"}
	case challenges.ImageChecking, challenges.ImageUnknown:
		return telemetry{0, "verificando"}
	case challenges.ImageNoDocker:
		return telemetry{-1, "sem docker"}
	case challenges.ImageDaemonDown:
		return telemetry{-1, "daemon parado"}
	default:
		return telemetry{-1, "build falhou"}
	}
}

func (m Model) telemetryLine(name string, t telemetry) string {
	var dot string
	switch t.ok {
	case 1:
		dot = GreenStyle.Render("●")
	case 0:
		dot = LabelStyle.Render(spinnerGlyph(m.frame))
	default:
		dot = LogErrorStyle.Render("✗")
	}
	return MutedStyle.Render(padRight(name, 7)) + dot + " " + TextStyle.Render(t.label)
}

// ---------------------------------------------------------------------------
// Painel principal (viewport) e painel de log
// ---------------------------------------------------------------------------

func (m Model) panelColor() lipgloss.Color {
	switch {
	case m.showHelp:
		return ColorCyanGlitch
	case m.screen == ScreenDecoder:
		return ColorAmber
	case m.screen == ScreenVictory:
		return ColorGreenCRT
	}
	if ch := m.currentChallenge(); ch != nil && m.Progress.IsCompleted(ch.ID) {
		return ColorGreenCRT
	}
	return ColorGreenDim
}

func (m Model) renderMainPanel(l layout) string {
	style := PanelStyle.BorderForeground(m.panelColor())
	cw := contentWidth(style, l.mainW)
	innerH := l.panelH - style.GetVerticalFrameSize()

	title := AccentStyle.Render(m.channelTitle())
	if m.glitchActive() {
		title = AccentStyle.Render(glitchText(m.channelTitle(), 6))
	}
	indicator := ""
	if m.viewport.TotalLineCount() > m.viewport.Height {
		arrows := "▼"
		switch {
		case m.viewport.AtBottom():
			arrows = "▲"
		case !m.viewport.AtTop():
			arrows = "▲▼"
		}
		indicator = DimStyle.Render(fmt.Sprintf("%s %3.0f%%", arrows, m.viewport.ScrollPercent()*100))
	}
	gap := max(cw-lipgloss.Width(title)-lipgloss.Width(indicator), 1)
	header := truncate(title+strings.Repeat(" ", gap)+indicator, cw)

	var body string
	if m.tuning > 0 {
		body = m.renderTuning(cw, m.viewport.Height)
	} else {
		body = m.viewport.View()
	}
	return style.Width(borderedWidth(style, l.mainW)).Height(innerH).MaxHeight(l.panelH).
		Render(header + "\n" + body)
}

// renderTuning mostra estática enquanto o canal é "sintonizado".
func (m Model) renderTuning(w, h int) string {
	noise := strings.Split(staticNoise(w, h), "\n")
	label := OnAirStyle.Render(fmt.Sprintf(" SINTONIZANDO CH %02d · %s ", int(m.screen)%10, m.channelName(m.screen)))
	if mid := h / 2; mid < len(noise) {
		noise[mid] = lipgloss.PlaceHorizontal(w, lipgloss.Center, label)
	}
	return strings.Join(noise, "\n")
}

func (m Model) renderLogPanel(w, h int) string {
	style := PanelStyle.BorderForeground(ColorGreenDim)
	cw := contentWidth(style, w)
	innerH := h - style.GetVerticalFrameSize()

	title := AccentStyle.Render("RECEPTOR DE SINAL · LOG AO VIVO") + DimStyle.Render(fmt.Sprintf("  (%d requisições)", len(m.logs)))
	lines := []string{truncate(title, cw)}
	lines = append(lines, m.logLines(cw, innerH-1)...)
	return style.Width(borderedWidth(style, w)).Height(innerH).MaxHeight(h).Render(strings.Join(lines, "\n"))
}

// logLines formata as últimas requisições que cabem em `rows` linhas.
func (m Model) logLines(cw, rows int) []string {
	if rows <= 0 {
		return nil
	}
	if !m.Server2.Running() {
		return []string{LogErrorStyle.Render(truncate("✗ Transmissor offline — nenhum sinal pode ser recebido.", cw))}
	}
	if len(m.logs) == 0 {
		return []string{DimStyle.Render(truncate(fmt.Sprintf("(aguardando sinal… em outro terminal: curl -i %s/)", m.Server2.URL()), cw))}
	}
	start := max(len(m.logs)-rows, 0)
	out := make([]string, 0, rows)
	for _, e := range m.logs[start:] {
		out = append(out, m.formatLogLine(e, cw))
	}
	return out
}

func (m Model) formatLogLine(e challenges.HTTPLogEvent, cw int) string {
	ts := DimStyle.Render(e.Timestamp.Format("15:04:05"))
	method := lipgloss.NewStyle().Bold(true).Foreground(methodColor(e.Method)).Render(padRight(e.Method, 6))
	status := lipgloss.NewStyle().Bold(true).Foreground(statusColor(e.StatusCode)).Render(fmt.Sprintf("%d", e.StatusCode))
	icon := DimStyle.Render("·")
	switch {
	case e.Step > 0:
		icon = GreenStyle.Render("★")
	case e.StatusCode >= 400:
		icon = LogErrorStyle.Render("✗")
	case e.Success:
		icon = GreenStyle.Render("✓")
	}
	prefix := ts + " " + icon + " " + method + " " + status + " "
	rest := cw - lipgloss.Width(prefix)
	pathW := min(max(rest/2, 12), 34)
	path := TextStyle.Render(padRight(truncate(e.Target(), pathW), pathW))
	msgW := rest - pathW - 1
	msg := MutedStyle.Render(truncate(e.Message, max(msgW, 0)))
	return truncate(prefix+path+" "+msg, cw)
}

// ---------------------------------------------------------------------------
// Rodapé (toast / letreiro EAS + atalhos)
// ---------------------------------------------------------------------------

func (m Model) renderFooter() string {
	var top string
	if m.toast != nil {
		style := lipgloss.NewStyle().Bold(true).Padding(0, 1)
		switch m.toast.kind {
		case toastSuccess:
			style = style.Foreground(ColorDarkBg).Background(ColorGreenCRT)
		case toastWarn:
			style = style.Foreground(ColorDarkBg).Background(ColorAmber)
		case toastError:
			style = style.Foreground(ColorWhite).Background(ColorRedAlert)
		default:
			style = style.Foreground(ColorDarkBg).Background(ColorCyanGlitch)
		}
		top = style.Render(truncate(m.toast.text, m.width-2))
	} else if m.cfg.NoGlitch {
		top = CrawlStyle.Render(truncate(strings.TrimSpace(crawlText), m.width))
	} else {
		top = CrawlStyle.Render(crawl(crawlText, m.width, m.frame))
	}

	h := m.help
	h.Width = m.width
	return top + "\n" + h.ShortHelpView(m.shortHelp())
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	mnt := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, mnt, s)
}
