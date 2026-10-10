package challenges

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ContingencyImage é a imagem Docker investigada na Transmissão 4.
const ContingencyImage = "local58-contingency:latest"

// SetupChallenge4 cria o Dockerfile de Contingência baseado no infame episódio
// 'Contingency' de Local 58 (Department of Preservation) e constrói a imagem
// de forma síncrona (usado pelo modo --setup-only).
func SetupChallenge4(workspaceDir string) error {
	if err := WriteChallenge4Files(workspaceDir); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_ = EnsureContingencyImage(ctx, workspaceDir)
	return nil
}

// Challenge4Dir devolve a pasta com o Dockerfile da contingência.
func Challenge4Dir(workspaceDir string) string {
	return filepath.Join(workspaceDir, "wclv_contingency")
}

// WriteChallenge4Files escreve o Dockerfile e o README da contingência (rápido,
// sem depender do Docker).
func WriteChallenge4Files(workspaceDir string) error {
	dockerDir := Challenge4Dir(workspaceDir)
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		return fmt.Errorf("falha ao criar pasta wclv_contingency: %w", err)
	}

	dockerfilePath := filepath.Join(dockerDir, "Dockerfile")
	dockerfileContent := fmt.Sprintf(`FROM alpine:latest
LABEL maintainer="DEPARTMENT OF PRESERVATION"
LABEL directive="ACT OF CONTINGENCY - SURRENDER TO THE NIGHT SKY"
LABEL protocol="PROCEDURE 3F: FRONT LAWN. FACE UP. FEET TOGETHER."

RUN echo "HOC VOLUNTAS TUA. O 51º ESTADO NÃO É UM LUGAR." > /contingency.txt

# CODIGO DE RESISTENCIA DO ULTIMO FUNCIONARIO ANTES DA RENDICAO
ENV PRESERVATION_OVERRIDE_CODE="%s"

# APAGANDO CODIGO DE RESISTENCIA - A RENDICAO E INEVITAVEL
ENV PRESERVATION_OVERRIDE_CODE=""

CMD ["sh", "-c", "echo 'ESTA MENSAGEM SE REPETIRÁ ATÉ QUE NÃO HAJA MAIS NINGUÉM PARA LÊ-LA.' && sleep infinity"]
`, Flag4)

	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("falha ao escrever Dockerfile: %w", err)
	}

	readmePath := filepath.Join(dockerDir, "README.txt")
	readmeContent := `[LOCAL 58 // DEPARTAMENTO DE PRESERVAÇÃO]
DIRETIVA DE CONTINGÊNCIA NACIONAL:
As forças no céu noturno são insuperáveis.
O último funcionário registrou o código de resistência na instrução
PRESERVATION_OVERRIDE_CODE antes de apagá-lo para cumprir a ordem de rendição.

As camadas de uma imagem Docker são eternas e imutáveis.
Inspecione as camadas da imagem local58-contingency:latest
para recuperar o código de resistência que sobreviveu à rendição.
`
	_ = os.WriteFile(readmePath, []byte(readmeContent), 0644)
	return nil
}

// ImageState representa a situação da imagem da contingência.
type ImageState int

const (
	ImageUnknown     ImageState = iota
	ImageChecking               // verificação/build em andamento
	ImageReady                  // imagem disponível localmente
	ImageNoDocker               // binário docker inexistente
	ImageDaemonDown             // docker instalado, mas daemon inacessível
	ImageBuildFailed            // build falhou
)

// ImageStatus é o diagnóstico da imagem Docker do desafio 4.
type ImageStatus struct {
	State  ImageState
	Detail string
}

// EnsureContingencyImage verifica se a imagem existe e, caso contrário, tenta
// construí-la. Nunca entra em pânico: sempre devolve um diagnóstico legível.
func EnsureContingencyImage(ctx context.Context, workspaceDir string) ImageStatus {
	if _, err := exec.LookPath("docker"); err != nil {
		return ImageStatus{State: ImageNoDocker, Detail: "binário 'docker' não encontrado no PATH"}
	}

	infoCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(infoCtx, "docker", "info", "--format", "{{.ServerVersion}}").CombinedOutput(); err != nil {
		return ImageStatus{State: ImageDaemonDown, Detail: lastLine(string(out), err)}
	}

	if exec.CommandContext(ctx, "docker", "image", "inspect", ContingencyImage).Run() == nil {
		return ImageStatus{State: ImageReady, Detail: "imagem disponível localmente"}
	}

	out, err := exec.CommandContext(ctx, "docker", "build", "-t", ContingencyImage, Challenge4Dir(workspaceDir)).CombinedOutput()
	if err != nil {
		return ImageStatus{State: ImageBuildFailed, Detail: lastLine(string(out), err)}
	}
	return ImageStatus{State: ImageReady, Detail: "imagem construída agora"}
}

func lastLine(out string, err error) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if l := strings.TrimSpace(lines[i]); l != "" {
			if len(l) > 160 {
				l = l[:157] + "..."
			}
			return l
		}
	}
	if err != nil {
		return err.Error()
	}
	return "erro desconhecido"
}
