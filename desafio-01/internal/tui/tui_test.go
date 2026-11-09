package tui

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/devops/desafio-analog-horror/internal/challenges"
)

func newTestModel(t *testing.T) (Model, *challenges.Progress, *challenges.ServerChallenge2, string) {
	tempDir, err := os.MkdirTemp("", "tui_test_*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}

	progress := challenges.NewProgress(tempDir)
	server := challenges.NewServerChallenge2(8097)

	cfg := Config{
		Progress:     progress,
		Server:       server,
		WorkspaceDir: tempDir,
		NoGlitch:     true,
	}

	m := NewModel(cfg)
	// Simula terminal 100x30
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model := updated.(Model)
	return model, progress, server, tempDir
}

func TestTUIModelInitAndViews(t *testing.T) {
	m, _, _, tempDir := newTestModel(t)
	defer os.RemoveAll(tempDir)

	if m.screen != ScreenCentral {
		t.Errorf("esperava tela inicial ScreenCentral, obteve %v", m.screen)
	}

	view := m.View()
	if !strings.Contains(view, "LOCAL 58") {
		t.Errorf("view inicial deve conter 'LOCAL 58'")
	}
	if !strings.Contains(view, "WCLV-TV") {
		t.Errorf("view inicial deve conter 'WCLV-TV'")
	}
	if !strings.Contains(view, "CH 00") {
		t.Errorf("view inicial deve referenciar canal CH 00")
	}

	// Testa tela pequena (deve renderizar aviso 'SINAL FRACO' sem pânico)
	smallModel, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	smallView := smallModel.(Model).View()
	if !strings.Contains(smallView, "SINAL FRACO") {
		t.Errorf("tela pequena deve exibir mensagem de 'SINAL FRACO'")
	}
}

func TestTUINavigation(t *testing.T) {
	m, _, _, tempDir := newTestModel(t)
	defer os.RemoveAll(tempDir)

	// Navega com número '1' para Desafio 1
	m1, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model1 := m1.(Model)
	if model1.screen != ScreenChallenge1 {
		t.Errorf("esperava ScreenChallenge1 após teclar '1', obteve %v", model1.screen)
	}
	view1 := model1.View()
	if !strings.Contains(view1, "AS FITAS CORROMPIDAS") {
		t.Errorf("CH 01 view deve conter o título do desafio 1")
	}

	// Navega com '2' para Desafio 2
	m2, _ := model1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model2 := m2.(Model)
	if model2.screen != ScreenChallenge2 {
		t.Errorf("esperava ScreenChallenge2 após teclar '2', obteve %v", model2.screen)
	}
	view2 := model2.View()
	if !strings.Contains(view2, "SKYWATCHING") {
		t.Errorf("CH 02 view deve conter 'SKYWATCHING'")
	}

	// Navega com 'f' para o Decodificador
	mDec, _ := model2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	modelDec := mDec.(Model)
	if modelDec.screen != ScreenDecoder {
		t.Errorf("esperava ScreenDecoder após teclar 'f', obteve %v", modelDec.screen)
	}
	viewDec := modelDec.View()
	if !strings.Contains(viewDec, "DECODIFICADOR") {
		t.Errorf("CH 05 view deve conter 'DECODIFICADOR'")
	}

	// Pressiona 'esc' para voltar ao canal anterior
	mBack, _ := modelDec.Update(tea.KeyMsg{Type: tea.KeyEsc})
	modelBack := mBack.(Model)
	if modelBack.screen != ScreenChallenge2 {
		t.Errorf("esperava retornar para ScreenChallenge2 após ESC, obteve %v", modelBack.screen)
	}

	// Testa ajuda com '?'
	mHelp, _ := modelBack.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	modelHelp := mHelp.(Model)
	if !modelHelp.showHelp {
		t.Errorf("esperava showHelp=true após teclar '?'")
	}
	viewHelp := modelHelp.View()
	if !strings.Contains(viewHelp, "MANUAL DO OPERADOR") {
		t.Errorf("view de ajuda deve conter 'MANUAL DO OPERADOR'")
	}
}

func TestTUIFlagSubmissionAndVictory(t *testing.T) {
	m, progress, _, tempDir := newTestModel(t)
	defer os.RemoveAll(tempDir)

	// Vai para o decodificador
	mDec, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	curModel := mDec.(Model)

	// Digita a Flag 1 no input
	curModel.input.SetValue(challenges.Flag1)
	mSub, _ := curModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	curModel = mSub.(Model)

	if !progress.IsCompleted(1) {
		t.Errorf("Flag 1 deveria ter sido completada")
	}
	if curModel.lastResult == nil || curModel.lastResult.Kind != challenges.SubmitCorrect {
		t.Errorf("resultado da submissão deveria ser SubmitCorrect")
	}

	// Submete as restantes
	curModel.input.SetValue(challenges.Flag2)
	mSub, _ = curModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	curModel = mSub.(Model)

	curModel.input.SetValue(challenges.Flag3)
	mSub, _ = curModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	curModel = mSub.(Model)

	curModel.input.SetValue(challenges.Flag4)
	mSub, _ = curModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	curModel = mSub.(Model)

	if !progress.IsAllCompleted() {
		t.Errorf("todas as flags deveriam estar completadas")
	}
	if curModel.screen != ScreenVictory {
		t.Errorf("tela deveria ser ScreenVictory após completar todas, obteve %v", curModel.screen)
	}

	vicView := curModel.View()
	if !strings.Contains(vicView, "TRANSMISSÃO FINALIZADA") {
		t.Errorf("view de vitória deve conter 'TRANSMISSÃO FINALIZADA'")
	}
}

func TestTUIEffects(t *testing.T) {
	// Testa colorBars
	bars := colorBars(80, false)
	if len(bars) == 0 {
		t.Errorf("colorBars não deve retornar vazio")
	}

	// Testa glitchText
	orig := "LOCAL 58 WCLV-TV"
	glitched := glitchText(orig, 2)
	if len(glitched) == 0 {
		t.Errorf("glitchText não deve retornar vazio")
	}

	// Testa crawl
	c := crawl("AVISO METEOROLOGICO", 30, 5)
	if len(c) == 0 {
		t.Errorf("crawl não deve retornar vazio")
	}

	// Testa segmentBar
	bar := segmentBar(2, 4)
	if !strings.Contains(bar, "▰") || !strings.Contains(bar, "▱") {
		t.Errorf("segmentBar deve conter caracteres preenchidos e vazios")
	}

	// Testa meterBar
	mBar := meterBar(0.5, 20, ColorGreenCRT)
	if len(mBar) == 0 {
		t.Errorf("meterBar não deve retornar vazio")
	}

	// Testa difficultyDots
	dots := difficultyDots(3)
	if !strings.Contains(dots, "●") || !strings.Contains(dots, "○") {
		t.Errorf("difficultyDots deve conter pontos cheios e vazios")
	}
}

func TestTUIHTTPLogAndToast(t *testing.T) {
	m, _, _, tempDir := newTestModel(t)
	defer os.RemoveAll(tempDir)

	evt := challenges.HTTPLogEvent{
		Timestamp:  time.Now(),
		Method:     "GET",
		Path:       "/weather-service",
		StatusCode: 302,
		Message:    "Sinal desviado para /skywatching",
		Success:    true,
		Step:       1,
	}

	mEvt, _ := m.Update(HTTPLogMsg(evt))
	updated := mEvt.(Model)

	if len(updated.logs) != 1 {
		t.Errorf("esperava 1 log registrado, obteve %d", len(updated.logs))
	}
	if updated.serverStep != 1 {
		t.Errorf("esperava serverStep == 1, obteve %d", updated.serverStep)
	}
	if updated.toast == nil {
		t.Errorf("esperava notificação toast após avanço de estágio")
	}
}

