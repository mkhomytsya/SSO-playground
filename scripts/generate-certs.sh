#!/usr/bin/env sh
#
# generate-certs.sh
# Creates a self-signed CA and a TLS server certificate for localhost.
# Outputs: /certs/server.crt, /certs/server.key, /certs/ca.crt
#
set -e

CERT_DIR="${CERT_DIR:-/certs}"
DAYS=365
SUBJ="/CN=SSO Playground CA"
SAN="DNS:localhost,DNS:oauth2-proxy,DNS:pocket-id,DNS:demo-app,IP:127.0.0.1"

# Ensure openssl is available
apk add --no-cache openssl >/dev/null 2>&1 || true

mkdir -p "$CERT_DIR"

# Skip if certs already exist and are still valid
if [ -f "$CERT_DIR/server.crt" ] && openssl x509 -checkend 86400 -noout -in "$CERT_DIR/server.crt" 2>/dev/null; then
  echo "Certificates already exist and are valid — skipping generation."
  exit 0
fi

echo "==> Generating self-signed CA …"
openssl req -x509 -nodes -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
  -keyout "$CERT_DIR/ca.key" -out "$CERT_DIR/ca.crt" \
  -days "$DAYS" -subj "$SUBJ"

echo "==> Generating server key + CSR …"
openssl req -nodes -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
  -keyout "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" \
  -subj "/CN=localhost"

echo "==> Signing server cert with CA …"
cat > "$CERT_DIR/extfile.cnf" <<EOF
subjectAltName=${SAN}
extendedKeyUsage=serverAuth
EOF

openssl x509 -req \
  -in "$CERT_DIR/server.csr" \
  -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial \
  -out "$CERT_DIR/server.crt" \
  -days "$DAYS" \
  -extfile "$CERT_DIR/extfile.cnf"

# Clean up intermediate files
rm -f "$CERT_DIR/server.csr" "$CERT_DIR/ca.key" "$CERT_DIR/ca.srl" "$CERT_DIR/extfile.cnf"

echo "==> Certificates written to $CERT_DIR"
ls -la "$CERT_DIR"
