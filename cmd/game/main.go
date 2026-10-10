package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/devops/desafio-analog-horror/internal/challenges"
	"github.com/devops/desafio-analog-horror/internal/tui"
)

func main() {
	workspaceDirFlag := flag.String("workspace", "", "Diretório de trabalho para os desafios (padrão: ./workspace)")
	portFlag := flag.Int("port", 8058, "Porta do servidor da transmissão de emergência (cURL)")
	setupOnlyFlag := flag.Bool("setup-only", false, "Apenas inicializa o workspace e desafios, sem abrir a TUI")
	serverOnlyFlag := flag.Bool("server-only", false, "Apenas executa o servidor HTTP na porta especificada, sem abrir a TUI")
	noGlitchFlag := flag.Bool("no-glitch", false, "Desativa efeitos visuais de estática analógica e glitch (acessibilidade)")
	flag.Parse()

	workspaceDir := *workspaceDirFlag
	if workspaceDir == "" {
		if envDir := os.Getenv("WORKSPACE_DIR"); envDir != "" {
			workspaceDir = envDir
		} else {
			workspaceDir = "./workspace"
		}
	}

	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao resolver caminho da workspace: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(absWorkspace, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar pasta da workspace: %v\n", err)
		os.Exit(1)
	}

	// 1. Inicializa Desafio 1 (Shell Maze)
	if err := challenges.SetupChallenge1(absWorkspace); err != nil {
		fmt.Fprintf(os.Stderr, "Aviso ao configurar Desafio 1: %v\n", err)
	}

	// 2. Inicializa Desafio 3 (Git Log Repo)
	if err := challenges.SetupChallenge3(absWorkspace); err != nil {
		fmt.Fprintf(os.Stderr, "Aviso ao configurar Desafio 3: %v\n", err)
	}

	// 3. Inicializa Desafio 4 (Docker Artifacts)
	if err := challenges.SetupChallenge4(absWorkspace); err != nil {
		fmt.Fprintf(os.Stderr, "Aviso ao configurar Desafio 4: %v\n", err)
	}

	if *setupOnlyFlag {
		fmt.Printf("Workspace inicializado com sucesso em %s!\n", absWorkspace)
		return
	}

	// 4. Inicializa o servidor do Desafio 2 (cURL API Relay)
	server2 := challenges.NewServerChallenge2(*portFlag)
	var serverErr error
	if err := server2.Start(); err != nil {
		serverErr = err
		fmt.Fprintf(os.Stderr, "Aviso: falha ao iniciar servidor HTTP na porta %d: %v\n", *portFlag, err)
		if *serverOnlyFlag {
			os.Exit(1)
		}
	}
	defer server2.Stop()

	if *serverOnlyFlag {
		fmt.Printf("[WCLV-TV] Servidor HTTP operando na porta %d (Ctrl+C para encerrar)...\n", *portFlag)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		for {
			select {
			case evt := <-server2.Events:
				fmt.Printf("[%s] %-6s %-24s -> %d (%s)\n",
					evt.Timestamp.Format("15:04:05"), evt.Method, evt.Target(), evt.StatusCode, evt.Message)
			case <-sigChan:
				fmt.Println("\n[WCLV-TV] Encerrando servidor HTTP...")
				return
			}
		}
	}

	// 5. Carrega o progresso
	progress := challenges.NewProgress(absWorkspace)

	// 6. Inicia o TUI Bubble Tea
	cfg := tui.Config{
		Progress:     progress,
		Server:       server2,
		WorkspaceDir: absWorkspace,
		ServerErr:    serverErr,
		NoGlitch:     *noGlitchFlag,
	}
	model := tui.NewModel(cfg)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar TUI: %v\n", err)
		os.Exit(1)
	}
}
