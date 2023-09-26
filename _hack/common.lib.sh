#!/bin/sh

DEFAULT_CLUSTER_NAME="operators-lab"

check_dep () {
    if ! command -v "${1}" >/dev/null 2>&1; then
        echo "No se encontró ${1}. Favor instalar e intentar de nuevo."
        return 1
    fi

    return 0
}

warn() {
    echo '\n===== ATENCION =====\n'
    echo ${1}
}