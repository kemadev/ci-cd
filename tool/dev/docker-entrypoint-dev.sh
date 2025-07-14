#!/usr/bin/env bash

set -euo pipefail

function main() {
    if [[ -z "${GO_MAIN_FILE_DIR}" ]]; then
        echo "GO_MAIN_FILE_DIR is not set."
        exit 1
    fi


    local SECRETS_DIR="/run/secrets"

    if [[ ! -d "${SECRETS_DIR}" ]]; then
        echo "Secrets directory '${SECRETS_DIR}' does not exist. Ensure that the secrets are mounted correctly."
        exit 1
    fi

    if [[ ! -f "${SECRETS_DIR}/netrc" ]]; then
        echo "The netrc file does not exist in '${SECRETS_DIR}'. Ensure that the secret is mounted correctly."
        exit 1
    fi

    cp /run/secrets/netrc ~/.netrc

    go build -o /app "${GO_MAIN_FILE_DIR}"

    exec /app "${@}" || {
        echo "Failure running the application, now sleeping."
        sleep inifinity
        exit 1
    }
}

main "${@}"
