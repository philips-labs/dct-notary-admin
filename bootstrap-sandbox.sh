#!/usr/bin/env sh

export NOTARY_ROOT_PASSPHRASE=t0pS3cr3t
export NOTARY_TARGETS_PASSPHRASE=t4rg3tr3p0
export NOTARY_SNAPSHOT_PASSPHRASE=sn4psh0t
export NOTARY_DELEGATION_PASSPHRASE=IcanSignImag3s
export NOTARY_AUTH=

export DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE=$NOTARY_ROOT_PASSPHRASE
export DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE=$NOTARY_TARGETS_PASSPHRASE

DOCKER_CONTENT_TRUST=0

REGISTRY=registry:5000

IMAGES="nginx:alpine alpine:latest busybox:latest traefik:latest"

docker trust key generate ci --dir ~/.docker/trust

for img in ${IMAGES}; do
    docker pull $img
    docker tag $img $REGISTRY/$img
    docker trust signer add continuous-integration --key ~/.docker/trust/ci.pub $REGISTRY/${img%:*}
    docker trust sign $REGISTRY/$img
done

unset NOTARY_ROOT_PASSPHRASE
unset NOTARY_TARGETS_PASSPHRASE
unset NOTARY_SNAPSHOT_PASSPHRASE
unset NOTARY_DELEGATION_PASSPHRASE
unset DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE
unset DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE
