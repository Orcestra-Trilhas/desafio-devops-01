package tui

import (
	"path/filepath"

	"github.com/devops/desafio-analog-horror/internal/challenges"
)

// Hint é um degrau da escada de dicas: o primeiro apenas aponta a direção,
// o segundo explica o conceito e o último entrega um caminho técnico.
type Hint struct {
	Text     string
	Commands []string
}

// Challenge descreve o conteúdo apresentado em cada canal de transmissão.
type Challenge struct {
	ID          int
	Short       string // nome curto para a grade de canais
	Title       string
	Skill       string
	Tools       string
	Difficulty  int
	Briefing    string // narrativa (lore)
	Objective   string // o que precisa ser feito, sem spoilers
	TargetLabel string
	Target      string // caminho, URL ou imagem a investigar
	Hints       []Hint
}

func buildCatalog(workspace, baseURL string) []Challenge {
	archive := filepath.Join(workspace, "wclv_archive")
	repo := filepath.Join(workspace, "wclv_station_repo")
	contingency := challenges.Challenge4Dir(workspace)
	img := challenges.ContingencyImage

	return []Challenge{
		{
			ID:         1,
			Short:      "FITAS",
			Title:      "AS FITAS CORROMPIDAS",
			Skill:      "SHELL",
			Tools:      "find · grep · xargs · pipes",
			Difficulty: 1,
			Briefing: "O acervo da emissora preserva décadas de gravações analógicas: episódios infantis macabros, " +
				"telemetria de veículos levados para a floresta e o feed óptico do telescópio. A fita de teste foi " +
				"regravada sobre mais de 15 setores com arquivos 'pass.txt' falsos, vazios ou cheios de ruído.",
			Objective:   "Encontre o único pass.txt autêntico do acervo e extraia a chave FLAG{...} escondida nas fitas.",
			TargetLabel: "ACERVO",
			Target:      archive,
			Hints: []Hint{
				{Text: "Nem todo pass.txt é igual: alguns estão vazios, outros só têm ruído ou chaves falsas. " +
					"Compare TAMANHO e CONTEÚDO, não apenas o nome do arquivo."},
				{Text: "O grep busca recursivamente (-r) e mostra arquivo e linha (-n). A âncora '^' exige que o " +
					"padrão comece no início da linha — isso descarta falsificações como FAKE_FLAG{...}. " +
					"O find também filtra por tamanho (-size +Nc)."},
				{Text: "Caminho técnico:", Commands: []string{
					`grep -rn "^FLAG{" ` + archive,
					`find ` + archive + ` -name pass.txt -size +10c -exec grep -H "^FLAG{" {} \;`,
				}},
			},
		},
		{
			ID:         2,
			Short:      "SKYWATCH",
			Title:      "O SINAL DE SKYWATCHING",
			Skill:      "HTTP/cURL",
			Tools:      "curl -i · -L · -H · -u · -X · -d",
			Difficulty: 2,
			Briefing: "O feed da WCLV está sendo sequestrado ao vivo pela transmissão de Skywatching. O transmissor " +
				"responde em JSON e deixa pistas nos cabeçalhos. Cada requisição sua aparece neste monitor em tempo real.",
			Objective:   "Conduza o foco do telescópio pelos 4 estágios do servidor até a Lua liberar a chave.",
			TargetLabel: "TRANSMISSOR",
			Target:      baseURL,
			Hints: []Hint{
				{Text: "Comece pela raiz: o servidor sempre responde com pistas. Use 'curl -i' para ver também o código " +
					"HTTP e os cabeçalhos X-Hint e X-Signal-Step. O INSPETOR abaixo mostra exatamente o que o servidor recebeu."},
				{Text: "Um 302 é um redirecionamento: o cURL só o segue com -L. Query strings vão na URL (?chave=valor, " +
					"sempre entre aspas no shell), cabeçalhos com -H 'Nome: valor', Basic Auth com -u usuario:senha e " +
					"corpo JSON com -d (que já implica POST)."},
				{Text: "Caminho técnico:", Commands: []string{
					"curl -iL " + baseURL + "/weather-service",
					"curl -H 'X-Transmission: 476MHz' '" + baseURL + "/skywatching?sky=visible'",
					"curl -u viewer:moon -H 'Content-Type: application/json' \\\n    -d '{\"action\": \"LOOK_AT_THE_MOON\"}' " + baseURL + "/rejoice",
				}},
			},
		},
		{
			ID:         3,
			Short:      "PLANTÃO",
			Title:      "O PLANTÃO DA MADRUGADA",
			Skill:      "GIT",
			Tools:      "git branch · git log · git show",
			Difficulty: 2,
			Briefing: "Na branch principal, o repositório da estação encerrou as transmissões à meia-noite. Mas em uma " +
				"linha do tempo paralela, os commits registram o operador enquanto a transmissão era tomada pelo aviso " +
				"meteorológico. Antes de abandonar o console e ir para o gramado, ele comitou as credenciais de emergência da torre.",
			Objective:   "Encontre a linha do tempo oculta e recupere a chave de corte da torre vazada no histórico.",
			TargetLabel: "REPOSITÓRIO",
			Target:      repo,
			Hints: []Hint{
				{Text: "A branch atual conta só a versão oficial dos fatos. Repositórios podem guardar linhas do tempo " +
					"paralelas que não aparecem no 'git log' padrão."},
				{Text: "Arquivos removidos continuam vivos no histórico. 'git log -p' mostra o diff de cada commit e " +
					"'--all' inclui todas as branches. 'git show <hash>' exibe um commit específico."},
				{Text: "Caminho técnico:", Commands: []string{
					"cd " + repo + " && git branch -a",
					"git log --oneline weather-service-hijack",
					"git log -p --all | grep -n 'FLAG{'",
				}},
			},
		},
		{
			ID:         4,
			Short:      "CONTING.",
			Title:      "O ATO DE CONTINGÊNCIA",
			Skill:      "DOCKER",
			Tools:      "docker image · docker history · docker inspect",
			Difficulty: 3,
			Briefing: "A imagem da transmissão de rendição da nação ao céu noturno foi construída pelo Departamento de " +
				"Preservação (Procedimento 3F: Gramado da Frente. Rosto para Cima. Pés Juntos). Durante o build, o último " +
				"operador registrou um código de resistência em uma variável de ambiente — e depois a apagou.",
			Objective:   "Recupere o código de resistência que sobreviveu nas camadas imutáveis da imagem " + img + ".",
			TargetLabel: "IMAGEM",
			Target:      img + "   (contexto: " + contingency + ")",
			Hints: []Hint{
				{Text: "Sobrescrever uma variável não apaga o passado: cada instrução do Dockerfile vira metadado ou " +
					"camada da imagem — e camadas são imutáveis."},
				{Text: "O Docker guarda as instruções que criaram cada camada. 'docker history' as lista, mas trunca " +
					"a saída por padrão; 'docker inspect' mostra apenas o estado FINAL da imagem."},
				{Text: "Caminho técnico:", Commands: []string{
					"docker history --no-trunc " + img,
					"docker history --no-trunc --format '{{.CreatedBy}}' " + img + " | grep PRESERVATION",
				}},
			},
		},
	}
}

