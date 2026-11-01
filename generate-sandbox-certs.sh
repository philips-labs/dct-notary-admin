#!/usr/bin/env bash
set -euo pipefail

# Development fixtures only. Keep the existing service keys, but replace the
# expired certificates with a fresh CA and mutually authenticated TLS chain.
fixtures=${1:?Usage: bash generate-sandbox-certs.sh <notary-fixtures-directory>}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

cat > "$work/openssl.cnf" <<'EOF'
[req]
distinguished_name = dn
[dn]
[ca]
basicConstraints = critical,CA:TRUE,pathlen:0
keyUsage = critical,keyCertSign,cRLSign
subjectKeyIdentifier = hash
[service]
basicConstraints = critical,CA:FALSE
keyUsage = critical,digitalSignature,keyEncipherment
extendedKeyUsage = serverAuth,clientAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
EOF

openssl genrsa -out "$work/root-ca.key" 2048
openssl req -new -x509 -sha256 -days 365 \
    -key "$work/root-ca.key" -subj '/CN=Notary Sandbox CA' \
    -config "$work/openssl.cnf" -extensions ca -out "$work/root-ca.crt"

for service in server signer; do
    name=notary-$service
    cp "$work/openssl.cnf" "$work/$name.cnf"
    printf 'subjectAltName = DNS:%s,DNS:notary%s,DNS:localhost,IP:127.0.0.1\n' \
        "$name" "$service" >> "$work/$name.cnf"
    openssl req -new -sha256 -key "$fixtures/$name.key" \
        -subj "/CN=$name" -out "$work/$name.csr"
    openssl x509 -req -sha256 -days 365 -in "$work/$name.csr" \
        -CA "$work/root-ca.crt" -CAkey "$work/root-ca.key" -CAcreateserial \
        -extfile "$work/$name.cnf" -extensions service -out "$work/$name.crt"
    openssl verify -CAfile "$work/root-ca.crt" -purpose sslserver \
        -verify_hostname "notary$service" "$work/$name.crt"
    openssl verify -CAfile "$work/root-ca.crt" -purpose sslclient "$work/$name.crt"
    # The signer trusts the server bundle as its client CA pool.
    cat "$work/root-ca.crt" >> "$work/$name.crt"
done

cp "$work/root-ca.crt" "$fixtures/root-ca.crt"
cp "$work/notary-server.crt" "$fixtures/notary-server.crt"
cp "$work/notary-signer.crt" "$fixtures/notary-signer.crt"
