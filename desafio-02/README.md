# Desafio 02 — Infraestrutura, Proxy e Resiliência

Desafio prático de troubleshooting e estabilização de infraestrutura conteinerizada.

Este desafio não envolve desenvolvimento de código de backend ou frontend. Todos os serviços utilizam imagens e ferramentas padrão: **Docker, Docker Compose, Nginx, PostgreSQL, Redis e GitHub Actions**.

---

## Arquitetura da Stack

```text
               Tráfego Externo
                      │
           ┌──────────┴──────────┐
           │                     │
      Porta 80 (HTTP)       Porta 443 (HTTPS)
           │                     │
           │ (Redireciona 301)   │
           └──────────► ┌────────▼────────┐
                        │   Nginx Proxy   │ (devops-proxy)
                        │ Terminação SSL  │
                        └────────┬────────┘
                                 │
                   Rede Interna (internal_net)
                                 │
         ┌───────────────────────┴───────────────────────┐
         │                                               │
┌────────▼────────┐                             ┌────────▼────────┐
│   api-node-1    │                             │   api-node-2    │ (Round-Robin)
└────────┬────────┘                             └────────┬────────┘
         │                                               │
         └───────────────────────┬───────────────────────┘
                                 │
                 ┌───────────────┴───────────────┐
                 │                               │
        ┌────────▼────────┐             ┌────────▼────────┐
        │   PostgreSQL    │             │      Redis      │
        │  (devops-postgres)│             │  (devops-redis) │
        │ Volume Persist. │             │   AOF Persist.  │
        └─────────────────┘             └─────────────────┘
```

---

## Contexto

A configuração inicial da stack apresenta problemas nas camadas de rede, proxy, certificados, healthchecks e persistência:

1. **Isolamento de rede:** O proxy Nginx não consegue alcançar os nós de aplicação devido à segregação de redes no Compose.
2. **Proxy reverso e upstream:** O bloco `upstream` aponta para portas incorretas e faltam cabeçalhos de proxy recomendados.
3. **HTTP -> HTTPS:** Requisições na porta 80 não realizam redirecionamento forçado para a porta 443.
4. **SSL/TLS:** O Nginx aponta para nomes/caminhos de certificado incorretos.
5. **Healthchecks:** As diretivas de verificação de saúde do PostgreSQL e do Redis falham ou utilizam comandos inadequados.
6. **Persistência:** O volume do banco está montado fora do diretório padrão de dados do PostgreSQL.
7. **Ambiente:** O arquivo `.env` precisa ser configurado a partir do `.env.example`.

O objetivo é identificar cada falha, corrigir os arquivos de configuração e validar a solução com o script de testes.

---

## Pré-requisitos

- Docker Engine 24+
- Docker Compose v2+
- OpenSSL e cURL

---

## Como Executar

Execute todos os comandos dentro do diretório `desafio-02`:

```bash
cd desafio-02
```

### 1. Gerar certificados SSL locais
```bash
./nginx/certs/generate-certs.sh
```

### 2. Configurar variáveis de ambiente
```bash
cp .env.example .env
```

### 3. Subir os containers
```bash
docker compose up -d
```

### 4. Executar os testes
```bash
./scripts/test.sh
```

---

## Critérios de Aceite

O script `./scripts/test.sh` valida 6 pontos da infraestrutura:

| Teste | Critério | Arquivos Relevantes |
| :--- | :--- | :--- |
| **1. Sintaxe e .env** | `docker compose config` valida sem erros com variáveis carregadas | `.env`, `docker-compose.yml` |
| **2. Healthchecks** | Todos os containers em execução com status `healthy` | `docker-compose.yml` |
| **3. Redirecionamento 80 -> 443** | Porta 80 responde com status 301/308 para URL HTTPS | `nginx/conf.d/default.conf` |
| **4. Conexão HTTPS** | Porta 443 responde 200 com certificados válidos e cabeçalhos de segurança | `nginx/conf.d/default.conf` |
| **5. Load Balancing** | Tráfego distribuído entre `api-node-1` e `api-node-2` | `docker-compose.yml`, `nginx/conf.d/default.conf` |
| **6. Banco e Cache** | PostgreSQL e Redis aceitam conexões e persistem dados | `docker-compose.yml`, `scripts/init-db.sql` |

---

## Comandos Úteis

- Inspecionar logs de um serviço:
  ```bash
  docker compose logs -f proxy
  docker compose logs -f database
  ```
- Verificar estado dos healthchecks:
  ```bash
  docker compose ps
  docker inspect --format='{{json .State.Health}}' devops-postgres
  ```
- Validar configuração do Nginx no container:
  ```bash
  docker compose exec proxy nginx -t
  docker compose exec proxy nginx -s reload
  ```
- Testar requisições com cabeçalhos:
  ```bash
  curl -Iv http://localhost/
  curl -Ivk https://localhost/
  ```

---

## Entrega

1. Garanta que todos os testes passem ao rodar `./scripts/test.sh`.
2. Documente as correções na [FOLHA_DE_RESPOSTA.md](FOLHA_DE_RESPOSTA.md).
3. Faça o commit dos arquivos alterados no repositório.
