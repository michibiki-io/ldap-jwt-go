#!/bin/bash
set -euo pipefail

map_uidgid() {
    USERMAP_ORIG_UID=$(id -u go)
    USERMAP_ORIG_GID=$(id -g go)
    USERMAP_GID=${USERMAP_GID:-$USERMAP_ORIG_GID}
    USERMAP_UID=${USERMAP_UID:-$USERMAP_ORIG_UID}
    if [ "${USERMAP_UID}" != "${USERMAP_ORIG_UID}" ] || [ "${USERMAP_GID}" != "${USERMAP_ORIG_GID}" ]; then
        echo "Starting with UID : $USERMAP_UID, GID: $USERMAP_GID"
        usermod -u $USERMAP_UID -o go
        groupmod -g $USERMAP_GID go
        chown go:go /opt/go
    fi
}

create_rsa() {

	if [ ! -e /opt/go/private/access.key ] && [ ! -e /opt/go/private/access.key.pub ]; then
		ssh-keygen -t rsa -b 4096 -f private/access.key -N "" -m PEM && ssh-keygen -f private/access.key.pub -e -m pkcs8 > private/access.key.pub.pkcs8
        rm -rf private/access.key.pub && mv private/access.key.pub.pkcs8 private/access.key.pub
        chown go:go /opt/go -R
        echo "RSA key pair for accessToken are generated."
	fi

	if [ ! -e /opt/go/private/access.key ] || [ ! -e /opt/go/private/access.key.pub ]; then
		echo "Could not found RSA keys, please check /opt/go/access.key or /opt/go/access.key.pub are exists."
        return 1
	fi

	if [ ! -e /opt/go/private/refresh.key ] && [ ! -e /opt/go/private/refresh.key.pub ]; then
		ssh-keygen -t rsa -b 4096 -f private/refresh.key -N "" -m PEM && ssh-keygen -f private/refresh.key.pub -e -m pkcs8 > private/refresh.key.pub.pkcs8
        rm -rf private/refresh.key.pub && mv private/refresh.key.pub.pkcs8 private/refresh.key.pub
        chown go:go /opt/go -R
        echo "RSA key pair for refreshToken are generated."
	fi

	if [ ! -e /opt/go/private/refresh.key ] || [ ! -e /opt/go/private/refresh.key.pub ]; then
		echo "Could not found RSA keys, please check /opt/go/refresh.key or /opt/go/refresh.key.pub are exists."
        return 1
	fi

    return 0

}

if [ "$(id -u)" = '0' ]; then
    create_rsa
    map_uidgid
    exec gosu go "$@"
else
    create_rsa
    exec "$@"
fi
