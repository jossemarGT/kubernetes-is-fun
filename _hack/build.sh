#!/bin/sh -e

. "$(dirname "${0}")/common.lib.sh"

build_image() {
    CONTEXT="${1}"
    IMAGE_NAME="$(basename $1)"
    VERSION="${2:-0.1.0}"
    NAMESPACE="kubernetes-is-fun"

    docker image build -f "${CONTEXT}/Dockerfile" -t "${NAMESPACE}/${IMAGE_NAME}:${VERSION}" "${CONTEXT}"
}

err_target() {
    >&2 echo "${1} Las opciones disponibles son:"
    find . -type f -name Dockerfile -exec dirname {} \; >&2 
    return 1
}

check_dep docker
check_dep minikube

if [ $# -lt 1 ]; then
    err_target 'No se seleccionó imagen a construir.'
fi

if [ ! -d "${1}" ] || [ ! -f "${1}/Dockerfile" ]; then
    err_target "Opción seleccionada \"${1}\" inválida."
fi

eval "$(minikube docker-env)"

for target in "${@}"; do
    build_image "${target}"
done
