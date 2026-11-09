# Folha de Resposta — Desafio DevOps 02

Utilize este documento para registrar as correções realizadas e o raciocínio aplicado em cada etapa do troubleshooting.

Para marcar uma opção, troque `[ ]` por `[x]`.

---

## Status da Bateria de Testes (`./scripts/test.sh`)

| Teste | Descrição | Status Final |
| :--- | :--- | :--- |
| **Teste 1** | Arquivo `.env` e Sintaxe do Compose | `[ ]` Passou &nbsp; `[ ]` Parcial &nbsp; `[ ]` Não passou |
| **Teste 2** | Status e Healthcheck dos Containers | `[ ]` Passou &nbsp; `[ ]` Parcial &nbsp; `[ ]` Não passou |
| **Teste 3** | Redirecionamento HTTP -> HTTPS (301/308) | `[ ]` Passou &nbsp; `[ ]` Parcial &nbsp; `[ ]` Não passou |
| **Teste 4** | Terminação SSL e Cabeçalhos de Segurança | `[ ]` Passou &nbsp; `[ ]` Parcial &nbsp; `[ ]` Não passou |
| **Teste 5** | Balanceamento de Carga (Load Balancing) | `[ ]` Passou &nbsp; `[ ]` Parcial &nbsp; `[ ]` Não passou |
| **Teste 6** | Conectividade e Persistência (Postgres/Redis) | `[ ]` Passou &nbsp; `[ ]` Parcial &nbsp; `[ ]` Não passou |

---

## Diagnóstico e Correções

### 1. Variáveis de Ambiente e Configuração Inicial
**Qual era a falha identificada?**



**O que precisou ser ajustado?**



---

### 2. Redes e Comunicação entre Containers (Docker Networks)
**Por que o Nginx não conseguia alcançar os serviços de aplicação (`api-node-1` e `api-node-2`)?**



**Quais alterações foram feitas no `docker-compose.yml` para corrigir a conectividade?**



---

### 3. Proxy Reverso e Balanceamento de Carga (Nginx)
**Quais inconsistências existiam no arquivo `nginx/conf.d/default.conf` em relação ao upstream?**



**Como o bloco `upstream` e os nós de backend foram configurados para distribuir as requisições?**



---

### 4. Certificados SSL/TLS e Redirecionamento HTTP -> HTTPS
**O que estava incorreto na configuração SSL inicial do Nginx?**



**Como foi implementado o redirecionamento de HTTP (porta 80) para HTTPS (porta 443)?**



**Quais cabeçalhos de segurança foram adicionados e qual a finalidade de cada um?**



---

### 5. Healthchecks e Ordem de Inicialização dos Serviços
**Por que os containers do PostgreSQL e do Redis falhavam nos healthchecks?**



**Quais comandos de teste foram configurados para validar a saúde de cada serviço?**

| Serviço | Comando Anterior | Comando Corrigido |
| :--- | :--- | :--- |
| PostgreSQL | | |
| Redis | | |

---

### 6. Persistência de Dados (Volumes)
**Qual era o erro no volume do PostgreSQL e qual o impacto caso o container fosse recriado?**



---

### 7. Pipeline de CI/CD (GitHub Actions)
**O que a pipeline valida e quais correções foram necessárias no arquivo de workflow?**



---

## Comandos Utilizados

Liste os comandos mais úteis utilizados durante o diagnóstico:

| Comando | Finalidade |
| :--- | :--- |
| | |
| | |
| | |
| | |

---

## Propostas de Melhoria para Produção

Quais melhorias de segurança, observabilidade ou arquitetura você recomendaria antes de colocar essa infraestrutura em produção real?



