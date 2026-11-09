#!/usr/bin/env bash
set -euo pipefail

CERTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
mkdir -p "$CERTS_DIR"

if [ -f "$CERTS_DIR/server.crt" ] && [ -f "$CERTS_DIR/server.key" ]; then
  echo "[✓] Certificados SSL já existem em $CERTS_DIR"
  exit 0
fi

echo "[*] Gerando certificados SSL autoassinados para localhost..."
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout "$CERTS_DIR/server.key" \
  -out "$CERTS_DIR/server.crt" \
  -subj "/C=BR/ST=SP/L=SaoPaulo/O=DevOps Challenge/OU=Infra/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

chmod 644 "$CERTS_DIR/server.crt"
chmod 600 "$CERTS_DIR/server.key"
echo "[✓] Certificados gerados com sucesso:"
echo "    - $CERTS_DIR/server.crt"
echo "    - $CERTS_DIR/server.key"
