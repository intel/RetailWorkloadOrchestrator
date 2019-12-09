#!/bin/bash
if [ ! -z ${PORT} ]; then
	glusterrest port ${PORT}
fi

if [ -z ${CREDS_DIR} ]; then
        # If CREDS_DIR is not provided, it defaults to /etc/rwo
        CREDS_DIR=/etc/rwo
fi

# Create creds DIR if it doesn't exist
mkdir -p ${CREDS_DIR}

# Delete previous users if any
users=`glusterrest show users | grep -v "User" | awk '{ print $1}'`
for user in ${users}
do
	glusterrest userdel ${user}
done

# Create a Random User
USER=`cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 10 | head -n 1`
echo $USER >  ${CREDS_DIR}/creds.txt

# Create a Random Password
PASSWORD=`cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 10 | head -n 1`
echo $PASSWORD >> ${CREDS_DIR}/creds.txt
chmod 400 ${CREDS_DIR}/creds.txt
glusterrest useradd -g glusteradmin -p ${PASSWORD} ${USER}
glusterrest usermod -g glusterroot ${USER}

glusterrestd "$@"
