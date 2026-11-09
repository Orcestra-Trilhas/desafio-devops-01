package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Paleta de cores Analog Horror / CRT / EAS
	ColorGreenCRT   = lipgloss.Color("#00FF66")
	ColorGreenDim   = lipgloss.Color("#008F39")
	ColorAmber      = lipgloss.Color("#FFB000")
	ColorAmberDim   = lipgloss.Color("#8A6200")
	ColorRedAlert   = lipgloss.Color("#FF3333")
	ColorRedDim     = lipgloss.Color("#7A1414")
	ColorDarkBg     = lipgloss.Color("#0D110D")
	ColorPanelBg    = lipgloss.Color("#111A11")
	ColorGrayDim    = lipgloss.Color("#556655")
	ColorGrayMid    = lipgloss.Color("#8FA38F")
	ColorCyanGlitch = lipgloss.Color("#00E5FF")
	ColorMagenta    = lipgloss.Color("#FF4FD8")
	ColorWhite      = lipgloss.Color("#E0FFE0")

	// Barras de cor SMPTE (padrão de teste de transmissão)
	SMPTEBars = []lipgloss.Color{
		"#C0C0C0", "#C0C000", "#00C0C0", "#00C000", "#C000C0", "#C00000", "#0000C0",
	}

	// Estilos Lipgloss
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreenCRT).
			Background(ColorDarkBg).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Italic(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGreenDim).
			Padding(0, 1)

	AlertBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorRedAlert).
			Padding(0, 1)

	WarnBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAmber).
			Padding(0, 1)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorGreenCRT).
			Padding(0, 1)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGreenDim).
			Padding(0, 1)

	MenuSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorDarkBg).
				Background(ColorGreenCRT)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(ColorGreenCRT)

	TagCompletedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorDarkBg).
				Background(ColorGreenCRT).
				Padding(0, 1)

	TagPendingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDarkBg).
			Background(ColorAmber).
			Padding(0, 1)

	TagAlertStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorRedAlert).
			Padding(0, 1)

	TagInfoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDarkBg).
			Background(ColorCyanGlitch).
			Padding(0, 1)

	LogSuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreenCRT).
			Bold(true)

	LogErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRedAlert).
			Bold(true)

	LogInfoStyle = lipgloss.NewStyle().
			Foreground(ColorAmber)

	CommandHintStyle = lipgloss.NewStyle().
				Foreground(ColorCyanGlitch)

	// Estilos de texto utilitários
	LabelStyle  = lipgloss.NewStyle().Bold(true).Foreground(ColorAmber)
	TextStyle   = lipgloss.NewStyle().Foreground(ColorWhite)
	DimStyle    = lipgloss.NewStyle().Foreground(ColorGrayDim)
	MutedStyle  = lipgloss.NewStyle().Foreground(ColorGrayMid)
	GreenStyle  = lipgloss.NewStyle().Foreground(ColorGreenCRT)
	AccentStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorGreenCRT)
	LoreStyle   = lipgloss.NewStyle().Foreground(ColorGrayMid).Italic(true)
	KeyStyle    = lipgloss.NewStyle().Bold(true).Foreground(ColorDarkBg).Background(ColorAmber).Padding(0, 1)

	// Cabeçalho / rodapé
	HeaderStyle = lipgloss.NewStyle().Background(ColorDarkBg).Foreground(ColorGreenCRT)
	OnAirStyle  = lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Background(ColorRedAlert).Padding(0, 1)
	OffAirStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorRedAlert).Background(ColorDarkBg).Padding(0, 1)
	CrawlStyle  = lipgloss.NewStyle().Foreground(ColorAmber).Background(ColorDarkBg)
	StaticStyle = lipgloss.NewStyle().Foreground(ColorGrayDim)
)

// borderedWidth devolve o valor a ser passado para Style.Width para que o
// bloco renderizado (incluindo bordas) ocupe exatamente `total` colunas.
func borderedWidth(s lipgloss.Style, total int) int {
	w := total - s.GetHorizontalBorderSize() - s.GetHorizontalMargins()
	if w < 1 {
		return 1
	}
	return w
}

// contentWidth devolve a largura útil de texto dentro do estilo.
func contentWidth(s lipgloss.Style, total int) int {
	w := total - s.GetHorizontalFrameSize()
	if w < 1 {
		return 1
	}
	return w
}

func statusColor(code int) lipgloss.Color {
	switch {
	case code >= 500:
		return ColorRedAlert
	case code >= 400:
		return ColorAmber
	case code >= 300:
		return ColorCyanGlitch
	case code >= 200:
		return ColorGreenCRT
	default:
		return ColorGrayMid
	}
}

func methodColor(method string) lipgloss.Color {
	switch method {
	case "GET", "HEAD":
		return ColorCyanGlitch
	case "POST":
		return ColorMagenta
	default:
		return ColorAmber
	}
}
