package tui

import (
	"time"

	"github.com/devops/desafio-analog-horror/internal/challenges"
)

// HTTPLogMsg transporta um evento do transmissor HTTP para a TUI.
type HTTPLogMsg challenges.HTTPLogEvent

// TickMsg é o relógio de animação (frames de estática, crawl, ON AIR).
type TickMsg time.Time

// ImageStatusMsg informa o resultado da verificação/build da imagem Docker.
type ImageStatusMsg challenges.ImageStatus

// FlagSubmitResultMsg é mantido por compatibilidade com integrações externas.
type FlagSubmitResultMsg struct {
	ChallengeID      int
	WasCorrect       bool
	AlreadyCompleted bool
	Message          string
}
