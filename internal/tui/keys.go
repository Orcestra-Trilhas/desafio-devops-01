package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap centraliza todos os atalhos do console. Os atalhos de rolagem não
// colidem com letras usadas por ações (h, f, r), e o mouse NÃO é capturado
// para que o jogador continue podendo selecionar/copiar texto do terminal.
type keyMap struct {
	NextCh   key.Binding
	PrevCh   key.Binding
	Jump     key.Binding
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Hint     key.Binding
	Decode   key.Binding
	Submit   key.Binding
	Rescan   key.Binding
	Back     key.Binding
	Help     key.Binding
	Quit     key.Binding
	ForceQ   key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		NextCh:   key.NewBinding(key.WithKeys("right", "tab", "l"), key.WithHelp("→/tab", "próx. canal")),
		PrevCh:   key.NewBinding(key.WithKeys("left", "shift+tab"), key.WithHelp("←/⇧tab", "canal ant.")),
		Jump:     key.NewBinding(key.WithKeys("0", "1", "2", "3", "4", "5"), key.WithHelp("0-5", "sintonizar")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "rolar")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "rolar")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("pgup", "página ↑")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+d", " "), key.WithHelp("pgdn", "página ↓")),
		Top:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("g/home", "topo")),
		Bottom:   key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("G/end", "fim")),
		Hint:     key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "revelar dica")),
		Decode:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "decodificar flag")),
		Submit:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "enviar")),
		Rescan:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "re-escanear ambiente")),
		Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "voltar")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "ajuda")),
		Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "desligar")),
		ForceQ:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "desligar")),
	}
}

// shortHelp devolve os atalhos relevantes para a tela atual.
func (m Model) shortHelp() []key.Binding {
	k := m.keys
	switch {
	case m.showHelp:
		return []key.Binding{k.Back, k.ForceQ}
	case m.screen == ScreenDecoder:
		return []key.Binding{k.Submit, k.Back, k.NextCh, k.ForceQ}
	case m.screen == ScreenVictory:
		return []key.Binding{k.Back, k.Jump, k.Quit}
	case m.currentChallenge() != nil:
		return []key.Binding{k.Hint, k.Decode, k.PrevCh, k.NextCh, k.Down, k.Help, k.Quit}
	default:
		return []key.Binding{k.Jump, k.NextCh, k.Decode, k.Rescan, k.Help, k.Quit}
	}
}

func (m Model) fullHelp() [][]key.Binding {
	k := m.keys
	return [][]key.Binding{
		{k.NextCh, k.PrevCh, k.Jump, k.Back},
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Top, k.Bottom},
		{k.Hint, k.Decode, k.Submit, k.Rescan},
		{k.Help, k.Quit, k.ForceQ},
	}
}

