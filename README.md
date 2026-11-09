# Desafios DevOps

Repositório contendo os desafios técnicos práticos de DevOps.

---

## Estrutura do Repositório

```text
.
├── desafio-01/          # Investigação de sistemas analógicos (Local 58)
├── desafio-02/          # Troubleshooting de infraestrutura (Docker, Nginx, SSL, Postgres, Redis)
└── .github/workflows/   # Workflows de CI
```

---

## Desafios

### [Desafio 01 — Investigação de Sistemas (Local 58)](desafio-01/README.md)
Desafio interativo baseado em TUI com investigação via terminal.

* **Tópicos:** Linux CLI (`find`, `grep`, permissões), HTTP/APIs (`curl`, headers, Base64), Git (branches, histórico, checkout) e Docker (inspeção de imagens).
* **Execução:**
  ```bash
  cd desafio-01
  ./run.sh
  ```
* **Folha de resposta:** [desafio-01/FOLHA_DE_RESPOSTA.md](desafio-01/FOLHA_DE_RESPOSTA.md)

---

### [Desafio 02 — Infraestrutura, Proxy e Resiliência](desafio-02/README.md)
Desafio prático de troubleshooting de infraestrutura conteinerizada (sem código de aplicação).

* **Tópicos:** Docker Compose v2, redes isoladas, Nginx como reverse proxy e load balancer, SSL/TLS com redirecionamento HTTP->HTTPS, healthchecks, persistência de volumes e CI com GitHub Actions.
* **Execução:**
  ```bash
  cd desafio-02
  ./nginx/certs/generate-certs.sh
  cp .env.example .env
  docker compose up -d
  ./scripts/test.sh
  ```
* **Folha de resposta:** [desafio-02/FOLHA_DE_RESPOSTA.md](desafio-02/FOLHA_DE_RESPOSTA.md)

---

## Entrega

1. Acesse o diretório do desafio desejado.
2. Siga as instruções do `README.md` correspondente.
3. Preencha a `FOLHA_DE_RESPOSTA.md` e faça o commit das alterações no repositório.
