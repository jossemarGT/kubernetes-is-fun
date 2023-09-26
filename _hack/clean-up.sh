#!/bin/sh -e

. "$(dirname "${0}")/common.lib.sh"

minikube delete --profile "${DEFAULT_CLUSTER_NAME}"
