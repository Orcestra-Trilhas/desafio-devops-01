#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "$DIR"

echo "============================================================"
echo " [WCLV-TV CANAL 58] TERMINAL DE TI // INVESTIGAÇÃO DEVOPS  "
echo "============================================================"

case "$1" in
  shell)
    echo "[*] Conectando ao terminal de investigação dentro do container..."
    docker exec -it desafio-devops-01 bash || docker compose exec desafio bash
    ;;
  local)
    shift || true
    echo "[*] Compilando binário Go..."
    go build -o desafio ./cmd/game
    echo "[*] Executando console localmente no host..."
    ./desafio --workspace=./workspace --port=8058 "$@"
    ;;
  build)
    echo "[*] Construindo imagens do Docker..."
    docker compose build
    ;;
  *)
    if ! docker info >/dev/null 2>&1; then
      echo "[!] Aviso: O serviço do Docker não está ativo na sua máquina."
      echo "[!] (Para ativar o Docker, execute em outro terminal: sudo systemctl start docker)"
      echo ""
      echo "[*] Alternando automaticamente para execução LOCAL direta..."
      echo ""
      exec "$0" local
    fi

    echo "[*] Iniciando desafio via Docker Compose..."
    echo "[*] DICA: Abra um SEGUNDO terminal e rode: ./run.sh shell"
    echo ""
    docker compose run --rm --service-ports desafio
    ;;
esac
