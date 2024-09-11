#!/bin/bash

TLSPATH=${1:-./tls}
SERVICENAME=${2:-registryreplace}
NAMESPACE=${3:-default}

# 生成 CA 私鑰和憑證
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -subj "/O=System:Nodes/CN=system:node:${SERVICENAME}.${NAMESPACE}.svc" -out ca.crt

# webhook 服務創建私鑰
openssl genrsa -out server.key 2048

# 創建憑證簽名請求 (CSR)。注意這裡添加了 Subject Alternative Name (SAN)
openssl req -new -key server.key -subj "/CN=${SERVICENAME}.${NAMESPACE}.svc" -out server.csr -config <(cat <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICENAME}.${NAMESPACE}.svc
DNS.2 = ${SERVICENAME}.${NAMESPACE}.svc.cluster.local
EOF
)

# 使用 CA 簽署 CSR 以生成服務器憑證
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650 -extensions v3_req -extfile <(cat <<EOF
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICENAME}.${NAMESPACE}.svc
DNS.2 = ${SERVICENAME}.${NAMESPACE}.svc.cluster.local
EOF
)

# Remove old tls files
rm -rf ${TLSPATH} && mkdir -p ${TLSPATH}
mv ca.crt ca.key ca.srl server.crt server.csr server.key ${TLSPATH}
