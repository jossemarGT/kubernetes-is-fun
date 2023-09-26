#!/bin/sh -e

. "$(dirname "${0}")/common.lib.sh"

check_dep minikube
check_dep docker

MK_EXTRA_FLAGS=''
if uname -s | grep -iq 'darwin'; then
    MK_EXTRA_FLAGS='--driver=docker'
fi

minikube start \
    $MK_EXTRA_FLAGS \
    --addons ingress \
    --kubernetes-version v1.27.6 \
    --profile "${DEFAULT_CLUSTER_NAME}"

minikube profile "${DEFAULT_CLUSTER_NAME}"

if uname -s | grep -iq 'darwin'; then
    warn "
    Recuerda usar el siguiente comando en otra terminal para poder acceder al cluster desde 127.0.0.1\n

    \n  minikube tunnel \n
    "
fi
