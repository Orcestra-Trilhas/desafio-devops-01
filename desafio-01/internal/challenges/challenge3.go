package challenges

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SetupChallenge3 cria o repositório Git em workspaceDir/wclv_station_repo
// onde a branch 'weather-service-hijack' documenta o sequestro da emissora
// pela transmissão da Lua em tempo real e o vazamento de chave pelo operador.
func SetupChallenge3(workspaceDir string) error {
	repoDir := filepath.Join(workspaceDir, "wclv_station_repo")

	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil
	}

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("falha ao criar pasta do repo: %w", err)
	}

	runGit := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Operador WCLV-TV",
			"GIT_AUTHOR_EMAIL=operator@local58.tv",
			"GIT_COMMITTER_NAME=Operador WCLV-TV",
			"GIT_COMMITTER_EMAIL=operator@local58.tv",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git %v falhou: %s: %w", args, string(out), err)
		}
		return nil
	}

	if err := runGit("init", "-b", "main"); err != nil {
		if err := runGit("init"); err != nil {
			return err
		}
		_ = runGit("checkout", "-b", "main")
	}

	_ = runGit("config", "user.name", "Operador WCLV-TV")
	_ = runGit("config", "user.email", "operator@local58.tv")

	// 1. Commits normais da estação na branch main
	_ = os.WriteFile(filepath.Join(repoDir, "station_profile.txt"), []byte("WCLV-TV • CANAL 58 • MASON COUNTY, WV\nLIDERANCA COMUNITARIA DESDE 1934\n"), 0644)
	_ = runGit("add", "station_profile.txt")
	_ = runGit("commit", "-m", "chore: cadastro institucional da emissora", "--date", "1999-10-30 18:00:00")

	_ = os.WriteFile(filepath.Join(repoDir, "sign_off.txt"), []byte("00:00 - Hino Nacional Americano\n00:05 - Encerramento das transmissões diárias\n00:06 - Estática e teste de barras\n"), 0644)
	_ = runGit("add", "sign_off.txt")
	_ = runGit("commit", "-m", "feat: rotina de encerramento da madrugada (sign-off)", "--date", "1999-10-30 23:55:00")

	// 2. Branch infectada pelo sinal da Lua
	branchName := "weather-service-hijack"
	if err := runGit("checkout", "-b", branchName); err != nil {
		return fmt.Errorf("falha ao criar branch %s: %w", branchName, err)
	}

	// Commit 1 na branch: Alerta inicial
	_ = os.WriteFile(filepath.Join(repoDir, "eas_alert.txt"), []byte("BOLETIM DE EMERGÊNCIA:\nUM EVENTO METEOROLÓGICO ESTÁ EM EFEITO.\nEVITEM TODOS OS ESPELHOS.\nNÃO OLHEM PARA CIMA.\n"), 0644)
	_ = runGit("add", "eas_alert.txt")
	_ = runGit("commit", "-m", "docs: ALERTA DO SERVICO METEOROLOGICO: NAO OLHEM PARA A LUA", "--date", "1999-10-31 02:45:00")

	// Commit 2 na branch: O sinal da Lua toma o texto
	_ = os.WriteFile(filepath.Join(repoDir, "eas_alert.txt"), []byte("AVISO RETIFICADO:\nO EVENTO METEOROLÓGICO ACABOU.\nÉ SEGURO SAIR.\nSAIAM DE SUAS CASAS AGORA.\nOLHEM PARA A LUA.\n"), 0644)
	_ = runGit("add", "eas_alert.txt")
	_ = runGit("commit", "-m", "fix: RETIFICACAO: E SEGURO SAIR. OLHEM PARA A LUA.", "--date", "1999-10-31 03:00:00")

	// Commit 3 na branch: Tentativa desesperada de corte com vazamento de chave
	secretContent := fmt.Sprintf("# CHAVE DE CORTE DA TORRE ANTES DE SUCUMBIR A LUA\nTOWER_OVERRIDE_KEY=%s\n", Flag3)
	_ = os.WriteFile(filepath.Join(repoDir, "tower_override.env"), []byte(secretContent), 0644)
	_ = runGit("add", "tower_override.env")
	_ = runGit("commit", "-m", "feat: chave de corte de transmissao da torre (remover antes do ar)", "--date", "1999-10-31 03:15:00")

	// Commit 4 na branch: Operador abandona o console
	_ = runGit("rm", "tower_override.env")
	_ = runGit("commit", "-m", "chore: o operador nao esta mais no console. gramado da frente.", "--date", "1999-10-31 03:33:00")

	// Commit 5 na branch: A mensagem final de Local 58
	_ = os.WriteFile(filepath.Join(repoDir, "eas_alert.txt"), []byte("ESTA MENSAGEM SE REPETIRÁ ATÉ QUE NÃO HAJA MAIS NINGUÉM PARA LÊ-LA.\n"), 0644)
	_ = runGit("add", "eas_alert.txt")
	_ = runGit("commit", "-m", "docs: ESTA MENSAGEM SE REPETIRA ATE QUE NAO HAJA MAIS NINGUEM PARA LE-LA", "--date", "1999-10-31 03:45:00")

	// 3. Volta para main limpa
	if err := runGit("checkout", "main"); err != nil {
		return fmt.Errorf("falha ao voltar para main: %w", err)
	}

	_ = os.WriteFile(filepath.Join(repoDir, "status.log"), []byte("1999-10-31 - Estação em repouso. Teste de barras em 476.000 MHz.\n"), 0644)
	_ = runGit("add", "status.log")
	_ = runGit("commit", "-m", "docs: programacao regular encerrada", "--date", "1999-10-31 04:00:00")

	return nil
}
