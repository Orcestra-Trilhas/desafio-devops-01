#!/usr/bin/env bash
# ==============================================================================
# Script de validacao de infraestrutura - Desafio DevOps 02
# ==============================================================================
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
CHALLENGE_DIR="$(cd "$SCRIPT_DIR/.." >/dev/null 2>&1 && pwd)"
cd "$CHALLENGE_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

TOTAL_TESTS=6
PASSED_TESTS=0

pass() {
  echo -e "  ${GREEN}[OK]${NC} $1"
  PASSED_TESTS=$((PASSED_TESTS + 1))
}

fail() {
  echo -e "  ${RED}[FAIL]${NC} $1"
  if [ -n "${2:-}" ]; then
    echo -e "    ${YELLOW}>${NC} $2"
  fi
}

echo -e "\n${BOLD}Validacao de Infraestrutura — Desafio DevOps 02${NC}\n"

# ------------------------------------------------------------------------------
# TESTE 1: Arquivo .env e sintaxe do Docker Compose
# ------------------------------------------------------------------------------
echo -e "${BOLD}[1/6] Variaveis de ambiente e sintaxe Compose...${NC}"
if [ ! -f .env ]; then
  fail "Arquivo .env nao encontrado." "Copie .env.example para .env e revise as variaveis."
else
  if docker compose config >/dev/null 2>&1; then
    pass "docker-compose.yml e .env validos."
  else
    ERR=$(docker compose config 2>&1 | head -n 3)
    fail "Falha ao validar docker compose config." "$ERR"
  fi
fi

# ------------------------------------------------------------------------------
# TESTE 2: Status e Saude dos Containers
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}[2/6] Status e healthchecks dos containers...${NC}"
EXPECTED_SERVICES=("devops-proxy" "devops-api-node-1" "devops-api-node-2" "devops-postgres" "devops-redis")
ALL_HEALTHY=true
FAILED_SERVICES=()

for svc in "${EXPECTED_SERVICES[@]}"; do
  STATUS=$(docker inspect --format='{{.State.Status}}' "$svc" 2>/dev/null | tr -d '\r\n' || true)
  if [ -z "$STATUS" ]; then
    ALL_HEALTHY=false
    FAILED_SERVICES+=("$svc (nao encontrado)")
  elif [ "$STATUS" != "running" ]; then
    ALL_HEALTHY=false
    FAILED_SERVICES+=("$svc (status: $STATUS)")
  else
    HEALTH=$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$svc" 2>/dev/null | tr -d '\r\n' || echo "unknown")
    if [ "$HEALTH" = "unhealthy" ]; then
      ALL_HEALTHY=false
      FAILED_SERVICES+=("$svc (unhealthy)")
    elif [ "$HEALTH" = "starting" ]; then
      ALL_HEALTHY=false
      FAILED_SERVICES+=("$svc (inicializando...)")
    fi
  fi
done

if [ "$ALL_HEALTHY" = true ]; then
  pass "Containers em execucao e com status healthy."
else
  fail "Containers com pendencias ou falha de healthcheck." "${FAILED_SERVICES[*]}"
fi

# ------------------------------------------------------------------------------
# TESTE 3: Redirecionamento HTTP -> HTTPS (Porta 80 -> 443)
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}[3/6] Redirecionamento HTTP (80) -> HTTPS (443)...${NC}"
HTTP_PORT=$(grep -E '^APP_PORT_HTTP=' .env 2>/dev/null | cut -d '=' -f2 | tr -d '\r\n' || echo "80")
HTTP_PORT=${HTTP_PORT:-80}

HTTP_RESP=$(curl -s -o /dev/null -w "%{http_code} %{redirect_url}" "http://localhost:${HTTP_PORT}/" 2>/dev/null || true)
HTTP_CODE=$(echo "$HTTP_RESP" | tr -d '\r\n' | awk '{print $1}')
REDIRECT_URL=$(echo "$HTTP_RESP" | tr -d '\r\n' | awk '{print $2}')

if [[ "$HTTP_CODE" =~ ^(301|308)$ ]] && [[ "$REDIRECT_URL" =~ ^https:// ]]; then
  pass "Porta $HTTP_PORT redirecionou com HTTP $HTTP_CODE para $REDIRECT_URL."
else
  fail "Porta $HTTP_PORT nao redirecionou para HTTPS (esperado 301/308)." "Status retornado: ${HTTP_CODE:-000} | Location: '${REDIRECT_URL:-vazio}'."
fi

# ------------------------------------------------------------------------------
# TESTE 4: Conexao HTTPS e Cabecalhos de Seguranca
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}[4/6] Terminacao SSL/TLS e cabecalhos de seguranca...${NC}"
HTTPS_PORT=$(grep -E '^APP_PORT_HTTPS=' .env 2>/dev/null | cut -d '=' -f2 | tr -d '\r\n' || echo "443")
HTTPS_PORT=${HTTPS_PORT:-443}

HTTPS_HEADERS=$(curl -k -s -I "https://localhost:${HTTPS_PORT}/" 2>/dev/null || true)
HTTPS_CODE=$(echo "$HTTPS_HEADERS" | head -n 1 | awk '{print $2}' | tr -d '\r\n')

if [ "$HTTPS_CODE" = "200" ]; then
  HAS_X_FRAME=$(echo "$HTTPS_HEADERS" | grep -i 'x-frame-options' || true)
  HAS_X_CONTENT=$(echo "$HTTPS_HEADERS" | grep -i 'x-content-type-options' || true)

  if [ -n "$HAS_X_FRAME" ] && [ -n "$HAS_X_CONTENT" ]; then
    pass "HTTPS respondeu 200 OK com cabecalhos X-Frame-Options e X-Content-Type-Options."
  else
    fail "HTTPS respondeu 200 OK mas faltam cabecalhos de seguranca." "Verifique X-Frame-Options e X-Content-Type-Options no Nginx."
  fi
else
  fail "Conexao HTTPS na porta $HTTPS_PORT nao retornou 200 OK." "Codigo retornado: ${HTTPS_CODE:-Falha de conexao}."
fi

# ------------------------------------------------------------------------------
# TESTE 5: Balanceamento de Carga entre api-node-1 e api-node-2
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}[5/6] Balanceamento de carga (api-node-1 e api-node-2)...${NC}"
NODE1_SEEN=false
NODE2_SEEN=false

for i in {1..8}; do
  BODY=$(curl -k -s "https://localhost:${HTTPS_PORT}/" 2>/dev/null || true)
  if echo "$BODY" | grep -q "api-node-1"; then
    NODE1_SEEN=true
  fi
  if echo "$BODY" | grep -q "api-node-2"; then
    NODE2_SEEN=true
  fi
done

if [ "$NODE1_SEEN" = true ] && [ "$NODE2_SEEN" = true ]; then
  pass "Requisicoes distribuidas entre api-node-1 e api-node-2."
else
  fail "Trafego nao alternou entre os nos." "api-node-1: $NODE1_SEEN | api-node-2: $NODE2_SEEN."
fi

# ------------------------------------------------------------------------------
# TESTE 6: Conectividade e Persistencia do Banco e Cache
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}[6/6] Servicos internos (PostgreSQL e Redis)...${NC}"
DB_USER=$(grep -E '^POSTGRES_USER=' .env 2>/dev/null | cut -d '=' -f2 | tr -d '\r\n' || echo "devops_admin")
DB_NAME=$(grep -E '^POSTGRES_DB=' .env 2>/dev/null | cut -d '=' -f2 | tr -d '\r\n' || echo "devops_production")

DB_READY=$(docker exec devops-postgres pg_isready -U "$DB_USER" -d "$DB_NAME" 2>/dev/null | tr -d '\r\n' || echo "failed")
REDIS_PONG=$(docker exec devops-redis redis-cli ping 2>/dev/null | tr -d '\r\n' || echo "failed")

if echo "$DB_READY" | grep -q "accepting connections" && [ "$REDIS_PONG" = "PONG" ]; then
  ROW_COUNT=$(docker exec -i devops-postgres psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM system_telemetry;" 2>/dev/null | tr -d ' \r\n' || echo "0")
  if [ "$ROW_COUNT" -gt 0 ]; then
    pass "PostgreSQL e Redis conectados com dados de telemetria inicializados ($ROW_COUNT registros)."
  else
    fail "PostgreSQL ativo mas tabela system_telemetry nao encontrada." "Verifique init-db.sql e o volume de dados."
  fi
else
  fail "Falha na comunicacao com PostgreSQL ou Redis." "Postgres: $DB_READY | Redis: $REDIS_PONG"
fi

# ------------------------------------------------------------------------------
# RESUMO FINAL
# ------------------------------------------------------------------------------
echo ""
if [ "$PASSED_TESTS" -eq "$TOTAL_TESTS" ]; then
  echo -e "${GREEN}Resultado: $TOTAL_TESTS/$TOTAL_TESTS testes passaram.${NC}\n"
  exit 0
else
  echo -e "${YELLOW}Resultado: $PASSED_TESTS/$TOTAL_TESTS testes passaram.${NC}\n"
  exit 1
fi
