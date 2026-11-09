package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/devops/desafio-analog-horror/internal/challenges"
)

func main() {
	workspaceDirFlag := flag.String("workspace", "./workspace", "Diretório de trabalho")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Uso: go run ./cmd/submit [-workspace=./workspace] <FLAG>\n")
		os.Exit(1)
	}

	p := challenges.NewProgress(*workspaceDirFlag)
	res := p.Submit(args[0])

	switch res.Kind {
	case challenges.SubmitCorrect:
		fmt.Printf("✓ SUCESSO: %s\n", res.Message)
	case challenges.SubmitAlready:
		fmt.Printf("◌ AVISO: %s\n", res.Message)
	case challenges.SubmitDecoy:
		fmt.Printf("✗ ANOMALIA: %s\n", res.Message)
		os.Exit(2)
	default:
		fmt.Printf("✗ FALHA: %s\n", res.Message)
		os.Exit(1)
	}
}

