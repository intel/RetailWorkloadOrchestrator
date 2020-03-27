#!/bin/bash
# CA keys are in pem format
# output keys are in pem frmat
#

[[ -z $1 ]] && echo "Usage: credtool_create_keys.sh <ca_key> <ca_certificate>" && exit


CAKEY=$1
CACERT=$2
KEYNAME="local"

create_key() {
	echo "Create Keys"
	openssl genrsa -out $1 4096
}

sign_cert() {
	echo "Signing Key"

	KEY=$1

	# Create CSR
	# this CN is comon name per host, this is used by allow
    HOSTNAME="localhost"
	openssl req -new -sha256 -key $1 -subj '/CN='${HOSTNAME} -out ${KEY}.csr
	# CSR Created

	# sign
	openssl x509 -req -days 360 -in ${KEY}.csr -CA ${CACERT} -CAkey ${CAKEY} -CAcreateserial -out ${KEY}.crt -sha256

	# check cert
	openssl x509 -text -noout -in ${KEY}.crt

    # convert to pem
    openssl x509 -outform pem -in ${KEY}.crt -out ${KEY}.crt.pem
}

generate_keys() {

	FOLDER=$2
	KEYNAME=$1

	mkdir -p ${FOLDER}
	create_key ${KEYNAME}.key
	sign_cert ${KEYNAME}.key

	rm ${KEYNAME}.key.crt && rm ${KEYNAME}.key.csr
	mv ${KEYNAME}.key.crt.pem ${KEYNAME}.pem

	mv ${KEYNAME}.key ${FOLDER}/
	mv ${KEYNAME}.pem ${FOLDER}/
	cp ${CACERT} ${FOLDER}/${KEYNAME}.ca
}

main() {
    echo "Generate Keys"

    echo "Keys for gluster"
    generate_keys glusterfs node_keys

    echo "Keys for glusterrestd"
    generate_keys glusterrestd node_keys

    echo "Keys for handler"
    generate_keys serfhandler node_keys

    echo "Keys for plugin"
    generate_keys plugin node_keys

}

main