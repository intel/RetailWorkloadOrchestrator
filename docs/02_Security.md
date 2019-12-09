## RWO Security Setup

`RWO` is a secure solution to form an automatic cluster of `EDGE` Nodes. It uses `serf` for node discovery. A symmetric key should be installed into each node at install time. Each node which has the same `serf` key will automatically form the cluster when `RWO` service is up.

In addition to `serf` for node discovery, `RWO` also sets up a replicated shared storage using `glusterfs`. The `gluster` server in `RWO` has [management encryption](https://access.redhat.com/documentation/en-us/red_hat_gluster_storage/3.4/html/administration_guide/ch23s04) enabled to make the communication secure. It uses [PKI](https://en.wikipedia.org/wiki/Public_key_infrastructure) based on RSA-4096.

`RWO` design uses a `docker inside docker` sandbox for applications which doesn't have access to the filesystem of the node. The applications and `docker swarm` are isolated from the host system and cannot access `keys` which are stored in the host filesystem under root user privilege.


### Introduction to RWO Security


- SERF uses a symmetric key `AES-256`
- Gluster uses `management encryption` with PKI
- Internel API server uses TLSv1.2.
- All keys are stored in the filesystem under root privilege.


<a name="serf-key"></a>
### Initial Serf Key

It is the most important key which is unique to a cluster of Nodes. It is `AES-256` based key which is kept in `keyring.json`.

Create a unique key and keep it in a json file at `/etc/ssl/rwo/keyring.json`.

`serf` can be used to create this key.

If `RWO` Service has been built we can make use of it to genrate this key.

```
 # docker run -t edge/serf:0.8.2 keygen
```

It will print key on the console, it can be added to a file `keyring.json`

Example of a `keyring.json`

```
[
  "NK4NP1gjvm2qDbWN3BRNw5Oc2D8r9VFJ9TDibcfwv1Y="
]
```

**Example:**

Suppose you have 3 nodes. i.e 3 computers running linux and docker.

These are the steps to be followed.

1. Create a file `keyring.json` with a unique key.

2. Copy the key at the path `/etc/ssl/rwo/keyring.json` on each of the nodes using a `usb-drive`. This key should be kept a secret and not be leaked.

3. Create PKI keys offline for each node and copy them at `/etc/ssl`. This is explained in the next section.


<a name="pki-keys"></a>
### PKI Keys

#### Introduction


`RWO` has another set of keys which are unique to each node. They are used by `gluster` and  `glusterfs API server`.

For each cluster, it is mandatory to create a `CA Key`, which is a `RSA-4096` key. This is unique to a cluster.

Then we create other `keys` and sign their `certificates` using this `CA Key`.

`Retail Workload Orchestrator` uses a distributed filesystem, `glusterfs` to maintain data persistence and availability over the cluster. The `glusterd` daemons running on all nodes use PKI to encrypt communication over network. However, to identify the members of the cluster, the PKI Certificates of the nodes, used by `glusterfs` need to be signed by the same CA(Certificate Authority). The CA can be local or an established one.


These keys are kept at `/etc/ssl`.


The list of PKI Keys needed is:

|Component | Key | Certificate | CACert |
| :-- | -- | --: | --: |
| *glusterfs-server* | glusterfs.key | glusterfs.pem | glusterfs.ca |
| *glusterfs REST API server* | glusterfsrestd.key | glusterfsrestd.pem | glusterfsrestd.ca |
| *docker plugin* | plugin.key | plugin.pem | plugin.ca |
| *serf handlers* | serfhandler.key | serf handler.pem | serfhandler.ca |

In the above table, the entries under `Certificate` column have signed public keys. They are to be signed by out unique `CA Key`.
The column `CACert` has the public key certficate of `CA Key`
The shell script to create these keys using openssl is shown in this document.

#### Steps
- create a `CA key` for the cluster,
- don't loose it
- Create the keys for RWO, sign then using `CA Key` and keep them in /etc/ssl

<a name="create-ca-key"></a>
#### Create CA Key

Here is a sample shell script which can be used to creater a `CA Key`. Copy the contents of the script into a file, say `credtool_create_ca.sh` and generate the `CA Key` as:

```
$ credtool_create_ca.sh sample_cluster_ca
```
A  `CA Key` and `CA Certificate` are created:

```
sample_cluster_ca.key
sample_cluster_ca.crt
```
Here is the script to create the `CA Key`.

NOTE: Don't forget the passphrase whihc is set while creating the key. It would be required during signing the TLS Keys.

```sh
#!/bin/bash
# CA keys are in pem format
# output keys are in pem frmat
#

[[ -z $1 ]] && echo "Usage: credtool_create_ca.sh <ca_name>" && exit

CA=$1
CAKEY=${CA}.key
CACERT=${CA}.crt

# Create a CA key
openssl genrsa -des3 -out ${CAKEY} 4096

# Create a CA Cert
openssl req -x509 -new -nodes -key ${CAKEY} -sha256 -days 1024 -out ${CACERT}

# Review a CA Cert
openssl x509 -in ${CACERT} -text

```

It is essential to protect and save the `CA Key` generated above. Don't keep it in the node. Use this key to generate the TLS Keys outside of node and then use a `usb drive` to copy the generated keys at the path '/etc/ssl'.

The script to create the `TLS Keys` is shown below. You can save the content of the script in a file, say `credtool_create_keys.sh`. 

```sh

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

```
<a name="create-tls-keys"></a>
**Step to generate** `TLS Keys`

1. [Create](#create-ca-key) a `CA Key` and `CA Cert`.
2. Use `credtool_create_keys.sh` and `CA Key/Cert` created in the above step to
generate TLS Keys for the node.

```
$ credtool_create_keys.sh sample_cluster_ca.key sample_cluster_ca.crt
```
The above steps will create a folder `node_keys` with all the keys in it. Now these keys can be copied into `/etc/ssl` as

```
# cp node_keys/* /etc/ssl/
```

Start `RWO` Service only after all keys are in place.

NOTE: Use the same `CA key` for generating the `TLS keys` for each Node. Generate separate `TLS keys` for each node. 


<a name="key-lifetimes"></a>
### Key Lifetimes

  * Serf: Admins need to install a new serf key and use it across the cluster after every three months once the node is installed in RWO cluster.
  * Docker: Admins need to rotate the new keys for docker ensuring that all the nodes are up in the cluster after every three months once the node is installed in the RWO cluster.

    Command for key rotation :

    `docker swarm ca --rotate`

  * Gluster: Admins need to install new keys every year after once the node is installed in the RWO cluster.

  **Incase of Key compromise**:- Admins need to generate and install new keys by bringing the particular node down.

  **Incase of CA Compromise**:- Admins need to generate and install new keys(Signed by new CA) by bringing down entire cluster.

### Key Rotation Strategies

In case the *admin* needs to change/rotate the keys, as per policy or in an event of compromise/leak.


#### Changing/Rotating `serf` keys


These are the steps to be followed.

1. Make sure that all nodes in the cluster are up
2. Login into *console* container/ or ssh into *console* container of **any node** and issue the following commnds:
	- `serf keygen`  - It will create a key(32 bytes) k2
	- `serf keys install <key k2>`
	- `serf keys use <key k2>` - It will make k2 as primary key.
	- `serf keys remove <key k1> [optional]` - It is upto the admin to remove the key or not.

It is advisabe to perform(**keygen/install/use**) operation when all nodes are up.

If any node is not update with the latest key for any reason, out of band mechanism to be used.

#### Changing/Rotating `TLS Keys`

If there is a need to change `TLS Keys` and it is known that cluster's `CA Key` is safe and not compromised, it is possible to [re-generate](#create-tls-keys) the `TLS Keys` for the node again.

NOTE: In case the `CA Key` is compromised, you need to [generate](#create-ca-key) a new `CA Key` and [regenerate](#create-tls-keys) `TLS keys` for all  nodes which are signe dby the new `CA key`.

## NOTE to Admin

The application can be deployed to the cluster by `admin` only, the admin can use [console](04_Deploy_Workloads.md#set-up-management-tools) or [swarm management UIs](04_Deploy_Workloads.md#set-up-management-tools).

_In the current version of `RWO`, we do not recommend to use the
volume mapping of `/var/lib/docker` in the `docker-compose.yml` of apps/stacks deployed on the swarm cluster._

