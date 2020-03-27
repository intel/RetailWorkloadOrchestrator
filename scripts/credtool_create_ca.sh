#!/bin/bash
# CA keys are in pem format
# output keys are in pem frmat
#

[[ -z $1 ]] && echo "Usage: credtool_create_ca.sh <ca_name>" && exit

CA=$1
CAKEY=${CA}.key
CACERT=${CA}.crt

# Create a CA key
openssl genrsa -out ${CAKEY} 4096

# Create a CA Cert
openssl req -x509 -new -nodes -key ${CAKEY} -sha256 -days 1024 -out ${CACERT} -subj "/C=US/ST=Arizona/L=Chandler/O=Security/CN=edge.io"

# Review a CA Cert
openssl x509 -in ${CACERT} -text
