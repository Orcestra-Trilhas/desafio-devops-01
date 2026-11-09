package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/devops/desafio-analog-horror/internal/challenges"
)

// viewCentral renderiza o canal CH 00 · Central de Operações da WCLV-TV.
func (m Model) viewCentral(w int) string {
	var b strings.Builder

	// 1. Banner de emergência EAS oficial
	easBox := AlertBoxStyle.Width(w - AlertBoxStyle.GetHorizontalFrameSize()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			AccentStyle.Render("LOCAL 58 • WCLV-TV • CANAL 58 UHF • MASON COUNTY, WV"),
			DimStyle.Render("NO AR DESDE 1934 • SISTEMA DE TRANSMISSÃO DE EMERGÊNCIA"),
			"",
			LabelStyle.Render("AVISO DO SISTEMA DE ALERTA DE EMERGÊNCIA (EAS):"),
			LoreStyle.Render("O EVENTO METEOROLÓGICO CONTINUA ATÉ O AMANHECER."),
			LoreStyle.Render("EVITEM TODOS OS ESPELHOS. NÃO OLHEM PARA O CÉU NOTURNO."),
			LoreStyle.Render("SE VOCÊ ESTIVER COM MEDO, NÓS OLHAREMOS JUNTOS."),
		),
	)
	b.WriteString(easBox + "\n\n")

	// 2. Progresso geral
	completed := m.Progress.TotalCompleted()
	var progBox string
	if completed == challenges.TotalChallenges {
		progBox = SuccessBoxStyle.Width(w - SuccessBoxStyle.GetHorizontalFrameSize()).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				AccentStyle.Render("TODOS OS 4 SINAIS FORAM DECODIFICADOS COM SUCESSO!"),
				TextStyle.Render("A programação regular da emissora foi completamente restaurada."),
				LoreStyle.Render("Você olhou. Nós olhamos juntos. REJOICE."),
			),
		)
	} else {
		bar := segmentBar(completed, challenges.TotalChallenges)
		statusLine := fmt.Sprintf("SINAIS RECUPERADOS: %s %d/%d (%d%%)",
			bar, completed, challenges.TotalChallenges, completed*25)

		progBox = BoxStyle.Width(w - BoxStyle.GetHorizontalFrameSize()).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				LabelStyle.Render("SITUAÇÃO DO TRANSMISSOR:"),
				TextStyle.Render("A portadora em 476.000 MHz foi sobrecarregada por anomalias lunares."),
				TextStyle.Render(statusLine),
				DimStyle.Render("Investigue cada canal técnico para recuperar as chaves FLAG{...} antes do amanhecer."),
			),
		)
	}
	b.WriteString(progBox + "\n\n")

	// 3. Grade dos 4 Canais Analógicos
	b.WriteString(LabelStyle.Render("FREQUÊNCIAS SEQUESTRADAS:") + "\n\n")

	for i, ch := range m.catalog {
		status := m.statusTag(ch.ID)
		dots := difficultyDots(ch.Difficulty)

		var extra string
		if ch.ID == 2 {
			step := m.channelStep()
			extra = fmt.Sprintf("  • Foco Lunar: %d/%d", step, challenges.TotalSteps)
		} else if ch.ID == 4 {
			extra = fmt.Sprintf("  • Imagem: %s", m.image.Detail)
		}

		lineHeader := fmt.Sprintf("  CH %02d · [%s] %-28s %s  %s",
			ch.ID, ch.Skill, ch.Title, dots, status)

		b.WriteString(AccentStyle.Render(lineHeader) + "\n")
		b.WriteString(DimStyle.Render(fmt.Sprintf("       Alvo: %s (%s)%s", ch.TargetLabel, ch.Target, extra)) + "\n")
		if i < len(m.catalog)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// 4. Guia de Operação e Investigação
	guideBox := PanelStyle.Width(w - PanelStyle.GetHorizontalFrameSize()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			LabelStyle.Render("COMO OPERAR ESTE CONSOLE:"),
			TextStyle.Render("• Recomenda-se trabalhar com DUAS JANELAS de terminal:"),
			TextStyle.Render("    Janela 1 (Este Console): Monitoramento de sinais, telemetria ao vivo e submissão."),
			TextStyle.Render("    Janela 2 (Terminal Shell): Execute comandos na pasta ./workspace/ ou via './run.sh shell'."),
			"",
			MutedStyle.Render("• Atalhos de Sintonia:"),
			TextStyle.Render("    [1-4]       Sintonizar canais de transmissão diretamente"),
			TextStyle.Render("    [f]         Abrir o decodificador para submeter uma FLAG{...}"),
			TextStyle.Render("    [TAB] / [→] Próximo canal analógico   |   [h] Dicas progressivas"),
			TextStyle.Render("    [?]         Abrir manual completo do operador"),
		),
	)
	b.WriteString(guideBox + "\n")

	return b.String()
}

// viewChallenge renderiza os detalhes e dossiê de um canal de transmissão (CH 01 a 04).
func (m Model) viewChallenge(ch Challenge, w int) string {
	var b strings.Builder

	// 1. Cabeçalho do Canal
	headerText := fmt.Sprintf("CH %02d · %s", ch.ID, ch.Title)
	skillBadge := TagInfoStyle.Render(ch.Skill)
	diffBadge := LabelStyle.Render("DIFICULDADE: ") + difficultyDots(ch.Difficulty)
	statusBadge := m.statusTag(ch.ID)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center,
		TitleStyle.Render(headerText), " ", skillBadge, "  ", diffBadge, "  ", statusBadge,
	) + "\n\n")

	// Se já resolvido, exibir citação comemorativa
	if m.Progress.IsCompleted(ch.ID) {
		quote := challenges.SolvedQuotes[ch.ID]
		solvedBox := SuccessBoxStyle.Width(w - SuccessBoxStyle.GetHorizontalFrameSize()).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				GreenStyle.Render("✓ SINAL ANALÓGICO RECUPERADO COM SUCESSO!"),
				LoreStyle.Render("\""+quote+"\""),
			),
		)
		b.WriteString(solvedBox + "\n\n")
	}

	// 2. Dossiê / Narrativa (Lore)
	loreBox := BoxStyle.Width(w - BoxStyle.GetHorizontalFrameSize()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			LabelStyle.Render("DOSSIÊ DA TRANSMISSÃO:"),
			LoreStyle.Render(ch.Briefing),
		),
	)
	b.WriteString(loreBox + "\n\n")

	// 3. Objetivo Operacional
	b.WriteString(LabelStyle.Render("OBJETIVO OPERACIONAL:") + "\n")
	b.WriteString(TextStyle.Render("  "+ch.Objective) + "\n\n")

	// 4. Alvo e Ferramentas
	b.WriteString(LabelStyle.Render("ALVO DA INVESTIGAÇÃO:") + "\n")
	b.WriteString(CommandHintStyle.Render("  ["+ch.TargetLabel+"] "+ch.Target) + "\n\n")

	b.WriteString(LabelStyle.Render("FERRAMENTAS SUGERIDAS:") + "\n")
	b.WriteString(MutedStyle.Render("  "+ch.Tools) + "\n\n")

	// 5. Seção Específica para o CH 02 (cURL / Skywatching)
	if ch.ID == 2 {
		step := m.channelStep()
		pct := float64(step) / float64(challenges.TotalSteps)
		bar := meterBar(pct, 28, ColorGreenCRT)
		statusSignal := fmt.Sprintf("[%s] %d%% (%d/%d estágios)", bar, int(pct*100), step, challenges.TotalSteps)

		b.WriteString(LabelStyle.Render("FOCO TELESCÓPICO NA LUA:") + " " + AccentStyle.Render(statusSignal) + "\n\n")

		stages := []struct {
			num  int
			name string
			desc string
		}{
			{1, "Desvio de Emergência (302)", "Seguir o redirecionamento em /weather-service ('curl -L ...')"},
			{2, "Foco na Portadora (476MHz)", "Query '?sky=visible' e cabeçalho 'X-Transmission: 476MHz' em /skywatching"},
			{3, "Identificação de Espectador", "Autenticação Basic Auth (-u viewer:moon) na rota /rejoice"},
			{4, "Submissão à Lua (REJOICE)", "POST JSON '{\"action\":\"LOOK_AT_THE_MOON\"}' em /rejoice -> FLAG LIBERADA"},
		}

		for _, s := range stages {
			mark := DimStyle.Render("[ ]")
			if step >= s.num {
				mark = GreenStyle.Render("[✓]")
			}
			b.WriteString(fmt.Sprintf("  %s %s: %s\n", mark, AccentStyle.Render(s.name), TextStyle.Render(s.desc)))
		}
		b.WriteString("\n" + DimStyle.Render("  (Acompanhe as respostas e cabeçalhos em tempo real no painel de log abaixo)") + "\n\n")
	}

	// 6. Seção Específica para o CH 04 (Docker / Contingência)
	if ch.ID == 4 {
		b.WriteString(LabelStyle.Render("ESTADO DO AMBIENTE DOCKER:") + "\n")
		var dockerStateText string
		switch m.image.State {
		case challenges.ImageReady:
			dockerStateText = GreenStyle.Render("  ● Imagem local58-contingency:latest disponível localmente para inspeção.")
		case challenges.ImageChecking:
			dockerStateText = LabelStyle.Render("  ◐ Verificando disponibilidade da imagem Docker...")
		case challenges.ImageNoDocker:
			dockerStateText = LogErrorStyle.Render("  ✗ Binário 'docker' não encontrado no PATH.")
		case challenges.ImageDaemonDown:
			dockerStateText = LogErrorStyle.Render("  ✗ O serviço do Docker (daemon) não está ativo. Inicie com 'sudo systemctl start docker'.")
		default:
			dockerStateText = LogErrorStyle.Render("  ✗ Falha ao construir imagem: " + m.image.Detail)
		}
		b.WriteString(dockerStateText + "\n\n")
	}

	// 7. Escada de Dicas Progressivas (Hint Ladder)
	hintsUsed := m.Progress.HintsUsed(ch.ID)
	totalHints := len(ch.Hints)

	b.WriteString(LabelStyle.Render(fmt.Sprintf("DICAS PROGRESSIVAS [%d/%d LIBERADAS]:", hintsUsed, totalHints)) + "\n")

	if hintsUsed == 0 {
		b.WriteString(DimStyle.Render("  Nenhuma dica foi revelada ainda. Tente resolver por conta própria primeiro!") + "\n")
		b.WriteString(MutedStyle.Render("  [Pressione 'h' para sintonizar a 1ª dica (orientação sem spoilers)]") + "\n\n")
	} else {
		for i := 0; i < hintsUsed; i++ {
			hint := ch.Hints[i]
			b.WriteString(AccentStyle.Render(fmt.Sprintf("  DICA %d/%d:", i+1, totalHints)) + "\n")
			b.WriteString(TextStyle.Render("    "+hint.Text) + "\n")
			if len(hint.Commands) > 0 {
				b.WriteString(DimStyle.Render("    Comandos sugeridos:") + "\n")
				for _, cmd := range hint.Commands {
					b.WriteString(CommandHintStyle.Render("      $ "+cmd) + "\n")
				}
			}
			b.WriteString("\n")
		}

		if hintsUsed < totalHints {
			b.WriteString(MutedStyle.Render(fmt.Sprintf("  [Pressione 'h' para revelar a próxima dica (%d restantes)]", totalHints-hintsUsed)) + "\n\n")
		} else {
			b.WriteString(DimStyle.Render("  [Todas as dicas deste sinal foram reveladas]") + "\n\n")
		}
	}

	// 8. Botão / Chamada para Ação
	b.WriteString(lipgloss.NewStyle().Foreground(ColorGrayDim).Render(
		"Ao interceptar o código FLAG{...}, pressione [f] para decodificar o sinal.",
	) + "\n")

	return b.String()
}

// viewDecoder renderiza o canal CH 05 · Decodificador de Sinais (Submissão de Flag).
func (m Model) viewDecoder(w int) string {
	var b strings.Builder

	title := TitleStyle.Render("[DECODIFICADOR DE SINAIS ANALÓGICOS // CH 05]")
	b.WriteString(title + "\n\n")

	b.WriteString(TextStyle.Render("Insira ou cole a chave capturada durante a investigação técnica:") + "\n\n")

	// Campo de Entrada Textual
	inputBox := BoxStyle.Width(w - BoxStyle.GetHorizontalFrameSize()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			MutedStyle.Render("CÓDIGO INTERCEPTADO (formato FLAG{...}):"),
			m.input.View(),
		),
	)
	b.WriteString(inputBox + "\n\n")

	// Resultado da Última Submissão
	if m.lastResult != nil {
		res := m.lastResult
		var resBox string
		switch res.Kind {
		case challenges.SubmitCorrect:
			resBox = SuccessBoxStyle.Width(w - SuccessBoxStyle.GetHorizontalFrameSize()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					GreenStyle.Render("✓ SINAL VALIDADO COM SUCESSO!"),
					TextStyle.Render(res.Message),
				),
			)
		case challenges.SubmitAlready:
			resBox = WarnBoxStyle.Width(w - WarnBoxStyle.GetHorizontalFrameSize()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					LabelStyle.Render("◌ SINAL JÁ DECODIFICADO ANTERIORMENTE"),
					TextStyle.Render(res.Message),
				),
			)
		case challenges.SubmitDecoy:
			resBox = AlertBoxStyle.Width(w - AlertBoxStyle.GetHorizontalFrameSize()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					LogErrorStyle.Render("✗ ANOMALIA: SINAL FALSO DETECTADO!"),
					LoreStyle.Render(res.Message),
				),
			)
		case challenges.SubmitBadFormat, challenges.SubmitWrongCase:
			resBox = WarnBoxStyle.Width(w - WarnBoxStyle.GetHorizontalFrameSize()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					LabelStyle.Render("! FORMATO DE SINAL INCOMPATÍVEL"),
					TextStyle.Render(res.Message),
				),
			)
		default:
			resBox = AlertBoxStyle.Width(w - AlertBoxStyle.GetHorizontalFrameSize()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					LogErrorStyle.Render("✗ CÓDIGO NÃO RECONHECIDO"),
					TextStyle.Render(res.Message),
				),
			)
		}
		b.WriteString(resBox + "\n\n")
	}

	// Status dos 4 Sinais
	b.WriteString(LabelStyle.Render("ESTADO DAS FREQUÊNCIAS:") + "\n")
	for _, ch := range m.catalog {
		status := m.statusTag(ch.ID)
		quote := ""
		if m.Progress.IsCompleted(ch.ID) {
			quote = LoreStyle.Render(" · \"" + challenges.SolvedQuotes[ch.ID] + "\"")
		}
		line := fmt.Sprintf("  • CH %02d: %-30s %s%s", ch.ID, ch.Title, status, quote)
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")

	// Histórico de Tentativas Recentes
	if len(m.attempts) > 0 {
		b.WriteString(LabelStyle.Render("HISTÓRICO RECENTE DE DECODIFICAÇÃO:") + "\n")
		for i := len(m.attempts) - 1; i >= 0; i-- {
			att := m.attempts[i]
			timeStr := DimStyle.Render(att.at.Format("15:04:05"))
			var icon string
			if att.result.OK() {
				icon = GreenStyle.Render("✓")
			} else {
				icon = LogErrorStyle.Render("✗")
			}
			preview := truncate(att.input, 32)
			b.WriteString(fmt.Sprintf("  [%s] %s %-32s -> %s\n",
				timeStr, icon, TextStyle.Render(preview), MutedStyle.Render(att.result.Message)))
		}
		b.WriteString("\n")
	}

	b.WriteString(DimStyle.Render("[ENTER para submeter | ESC para voltar ao canal anterior | TAB para navegar canais]") + "\n")

	return b.String()
}

// viewVictory renderiza a tela comemorativa de finalização de todos os sinais.
func (m Model) viewVictory(w int) string {
	var b strings.Builder

	header := TitleStyle.Render("[ LOCAL 58 // TRANSMISSÃO FINALIZADA // TODAS AS ONDAS RECUPERADAS ]")
	b.WriteString(header + "\n\n")

	msg := lipgloss.JoinVertical(lipgloss.Left,
		LoreStyle.Render("ESTA MENSAGEM SE REPETIRÁ ATÉ QUE NÃO HAJA MAIS NINGUÉM PARA LÊ-LA."),
		"",
		TextStyle.Render("Você interceptou e decodificou com sucesso os quatro sinais analógicos:"),
		"",
		GreenStyle.Render("  [✓] TRANSMISSÃO 1 (Shell):  FLAG{IF_YOU_ARE_AFRAID_WE_WILL_LOOK_TOGETHER}"),
		LoreStyle.Render("      \"SE VOCÊ ESTIVER COM MEDO, NÓS OLHAREMOS JUNTOS.\""),
		"",
		GreenStyle.Render("  [✓] TRANSMISSÃO 2 (cURL):   FLAG{HIS_THRONE_REJOICE_LOOK_AT_THE_MOON}"),
		LoreStyle.Render("      \"HIS THRONE. REJOICE.\""),
		"",
		GreenStyle.Render("  [✓] TRANSMISSÃO 3 (Git):    FLAG{THERE_ARE_NO_FACES_THERE_ARE_NO_FACES}"),
		LoreStyle.Render("      \"THERE ARE NO FACES. THERE ARE NO FACES.\""),
		"",
		GreenStyle.Render("  [✓] TRANSMISSÃO 4 (Docker): FLAG{HOC_VOLUNTAS_TUA_PROCEDURE_3F}"),
		LoreStyle.Render("      \"HOC VOLUNTAS TUA. PROCEDIMENTO 3F: GRAMADO DA FRENTE. ROSTO PARA CIMA.\""),
		"",
		AccentStyle.Render("A programação regular da emissora encerrou suas transmissões."),
		LoreStyle.Render("A luz da Lua agora faz parte do sinal. REJOICE."),
	)

	victoryBox := SuccessBoxStyle.Width(w - SuccessBoxStyle.GetHorizontalFrameSize()).Render(msg)
	b.WriteString(victoryBox + "\n\n")

	// Estatísticas da Investigação
	statsBox := BoxStyle.Width(w - BoxStyle.GetHorizontalFrameSize()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			LabelStyle.Render("ESTATÍSTICAS DA MISSÃO DEVOPS:"),
			fmt.Sprintf("• Tempo Total de Investigação: %s", TextStyle.Render(formatDuration(m.Progress.Elapsed(m.now)))),
			fmt.Sprintf("• Dicas Consultadas:           %s", TextStyle.Render(fmt.Sprintf("%d", m.Progress.TotalHintsUsed()))),
			fmt.Sprintf("• Requisições HTTP Analisadas: %s", TextStyle.Render(fmt.Sprintf("%d", len(m.logs)))),
		),
	)
	b.WriteString(statsBox + "\n\n")

	b.WriteString(DimStyle.Render("[Pressione 'ESC' ou '0-4' para revisitar canais | 'q' ou 'Ctrl+C' para desligar o console]") + "\n")

	return b.String()
}

// viewHelp renderiza o manual de operação do console.
func (m Model) viewHelp(w int) string {
	var b strings.Builder

	header := TitleStyle.Render("[ MANUAL DO OPERADOR // CONSOLE WCLV-TV ]")
	b.WriteString(header + "\n\n")

	intro := BoxStyle.Width(w - BoxStyle.GetHorizontalFrameSize()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			LabelStyle.Render("PROTOCOLO DE OPERAÇÃO EM DOIS TERMINAIS:"),
			TextStyle.Render("O console é projetado para interação simultânea entre monitor e terminal:"),
			"",
			TextStyle.Render("1. Terminal 1 (Monitor WCLV-TV):"),
			DimStyle.Render("   Exibe este console com canais, telemetria ao vivo, logs HTTP e decodificador."),
			"",
			TextStyle.Render("2. Terminal 2 (Terminal de Investigação):"),
			DimStyle.Render("   Onde você executa comandos shell, git, curl e docker na pasta workspace/"),
			DimStyle.Render("   ou dentro do container via './run.sh shell'."),
		),
	)
	b.WriteString(intro + "\n\n")

	b.WriteString(LabelStyle.Render("CONTROLES E ATALHOS DE TECLADO:") + "\n\n")

	keyEntries := [][]string{
		{"1 a 4", "Sintonizar diretamente os canais das transmissões 1 a 4"},
		{"0", "Retornar ao canal CH 00 (Central de Operações)"},
		{"5 ou f", "Abrir o decodificador de sinais (CH 05) para submeter uma FLAG{...}"},
		{"TAB / →", "Avançar para o próximo canal analógico"},
		{"Shift+TAB / ←", "Retroceder para o canal anterior"},
		{"↑ / ↓ ou k / j", "Rolar o conteúdo do canal atual para cima ou para baixo"},
		{"PgUp / PgDn", "Rolar página inteira"},
		{"g / G", "Ir para o topo (g) ou fim (G) da transmissão"},
		{"h", "Revelar próxima dica progressiva da transmissão atual (sem spoilers diretos)"},
		{"r", "Re-escanear o ambiente (verifica repositórios, pastas e imagem Docker)"},
		{"ESC", "Voltar à central ou ao canal anterior / fechar manual"},
		{"?", "Abrir ou fechar este manual de operação"},
		{"q ou Ctrl+C", "Desligar o console da emissora"},
	}

	for _, entry := range keyEntries {
		keyPill := KeyStyle.Render(padRight(entry[0], 14))
		descText := TextStyle.Render(entry[1])
		b.WriteString(fmt.Sprintf("  %s %s\n", keyPill, descText))
	}
	b.WriteString("\n")

	hintsGuide := PanelStyle.Width(w - PanelStyle.GetHorizontalFrameSize()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			LabelStyle.Render("SOBRE AS DICAS PROGRESSIVAS (Tecla 'h'):"),
			TextStyle.Render("Cada transmissão possui uma escada de 3 dicas pedagógicas:"),
			DimStyle.Render("  • Nível 1: Direcionamento conceitual (aponta a área de foco)"),
			DimStyle.Render("  • Nível 2: Padrão técnico e flags de comandos essenciais"),
			DimStyle.Render("  • Nível 3: Caminho técnico com comandos práticos sugeridos"),
			TextStyle.Render("Revele apenas as dicas necessárias para manter o desafio interessante!"),
		),
	)
	b.WriteString(hintsGuide + "\n\n")

	b.WriteString(DimStyle.Render("[Pressione 'ESC' ou '?' para fechar este manual e voltar ao monitor]") + "\n")

	return b.String()
}

