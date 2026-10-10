package tui

import (
	"math/rand/v2"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Efeitos visuais de "analog horror": barras SMPTE, estática, glitch e crawl.

var noiseRunes = []rune(" ░▒▓█")
var glitchRunes = []rune("▓▒░█▀▄#%&@$?!/\\|")

// colorBars desenha uma faixa de barras de teste SMPTE com `width` colunas.
// Quando glitch=true, a faixa sofre um deslocamento horizontal (tracking error).
func colorBars(width int, glitch bool) string {
	if width <= 0 {
		return ""
	}
	offset := 0
	if glitch {
		offset = rand.IntN(width/4 + 1)
	}
	var b strings.Builder
	n := len(SMPTEBars)
	runStart, runIdx := 0, -1
	flush := func(end int) {
		if runIdx < 0 || end <= runStart {
			return
		}
		b.WriteString(lipgloss.NewStyle().Background(SMPTEBars[runIdx]).Render(strings.Repeat(" ", end-runStart)))
	}
	for x := 0; x < width; x++ {
		idx := ((x + offset) * n / width) % n
		if glitch && rand.IntN(18) == 0 {
			flush(x)
			runIdx = -1
			b.WriteString(StaticStyle.Render(string(noiseRunes[rand.IntN(len(noiseRunes))])))
			runStart = x + 1
			continue
		}
		if idx != runIdx {
			flush(x)
			runStart, runIdx = x, idx
		}
	}
	flush(width)
	return b.String()
}

// staticNoise gera um bloco de estática analógica (w x h).
func staticNoise(w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	lines := make([]string, h)
	buf := make([]rune, w)
	for y := 0; y < h; y++ {
		for x := range buf {
			buf[x] = noiseRunes[rand.IntN(len(noiseRunes))]
		}
		lines[y] = string(buf)
	}
	return StaticStyle.Render(strings.Join(lines, "\n"))
}

// glitchText troca aleatoriamente alguns caracteres por ruído.
func glitchText(s string, chance int) string {
	if chance <= 0 {
		return s
	}
	r := []rune(s)
	for i := range r {
		if r[i] != ' ' && rand.IntN(chance) == 0 {
			r[i] = glitchRunes[rand.IntN(len(glitchRunes))]
		}
	}
	return string(r)
}

// crawl devolve uma janela de `width` colunas de um texto rolando (letreiro EAS).
func crawl(text string, width, offset int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(text)
	if len(r) == 0 {
		return strings.Repeat(" ", width)
	}
	out := make([]rune, width)
	for i := 0; i < width; i++ {
		out[i] = r[(offset+i)%len(r)]
	}
	return string(out)
}

// segmentBar desenha uma barra segmentada (▰▱) com `total` segmentos.
func segmentBar(done, total int) string {
	if total <= 0 {
		return ""
	}
	if done > total {
		done = total
	}
	return GreenStyle.Render(strings.Repeat("▰", done)) + DimStyle.Render(strings.Repeat("▱", total-done))
}

// meterBar desenha uma barra contínua com `width` colunas para a fração 0..1.
func meterBar(frac float64, width int, color lipgloss.Color) string {
	if width <= 0 {
		return ""
	}
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(frac*float64(width) + 0.5)
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		DimStyle.Render(strings.Repeat("░", width-filled))
}

// difficultyDots representa a dificuldade em 4 níveis.
func difficultyDots(level int) string {
	return LabelStyle.Render(strings.Repeat("●", level)) + DimStyle.Render(strings.Repeat("○", 4-level))
}

var spinnerFrames = []string{"◐", "◓", "◑", "◒"}

func spinnerGlyph(frame int) string { return spinnerFrames[frame%len(spinnerFrames)] }

// truncate corta uma string (sensível a largura visual) com reticências.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > width {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// padRight completa com espaços até `width` colunas visuais.
func padRight(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}
