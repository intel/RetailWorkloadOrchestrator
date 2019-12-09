#!/bin/sh
if [ ! -f /etc/ssh/ssh_host_rsa_key ]; then
	ssh-keygen -f /etc/ssh/ssh_host_rsa_key -N "" -t rsa -b 4096
fi
if [ ! -f /etc/ssh/ssh_host_dsa_key ]; then
	ssh-keygen -f /etc/ssh/ssh_host_dsa_key -N "" -t dsa
fi
if [ ! -f /etc/ssh/ssh_host_ecdsa_key ]; then
	ssh-keygen -f /etc/ssh/ssh_host_ecdsa_key -N "" -t ecdsa -b 512
fi
if [ ! -f /etc/ssh/ssh_host_ed25519_key ]; then
	ssh-keygen -f /etc/ssh/ssh_host_ed25519_key -N "" -t ed25519
fi
echo "alias ls='ls --color=always'" >> /etc/profile
if [ ! -z ${PATH} ]; then
	echo "export PATH=${PATH}" >> /etc/profile
fi

if [ ! -z ${PORT} ]; then
	echo "Port ${PORT}" >> /etc/ssh/sshd_config
fi

if [ ! -z ${USERNAME} ] && [ ! -z ${PASSWORD} ]; then
	echo -e "${PASSWORD}\n${PASSWORD}\n" | adduser ${USERNAME} &&
	echo "${USERNAME} ALL=(ALL) ALL" >> /etc/sudoers
fi

if [ ! -z ${AUTHORIZED_KEYS} ] && [ ! -z ${USERNAME} ]; then
	echo -e "${PASSWORD}\n${PASSWORD}\n" | adduser ${USERNAME} -s /bin/bash &&
	echo "${USERNAME} ALL=(ALL) ALL" >> /etc/sudoers
	mkdir -p /home/${USERNAME}/.ssh &&
	chmod 0700 /home/${USERNAME}/.ssh &&
	echo "${AUTHORIZED_KEYS}" > /home/${USERNAME}/.ssh/authorized_keys
fi

if [ ! -z ${AUTHORIZED_KEYS} ] || [ ! -z ${USERNAME} ]; then
	/usr/sbin/sshd -D
else
	tail -f /dev/null
fi
