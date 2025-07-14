#!/usr/bin/env sh

set -eu

main() {
    if [ -z "${GO_MAIN_FILE_DIR:-}" ]; then
        echo "GO_MAIN_FILE_DIR is not set."
        exit 1
    fi

    cp /run/secrets/netrc ~/.netrc

    go build -o /app "${GO_MAIN_FILE_DIR}"

    set +e
    /app "${@}"
    EXIT_CODE=$?
    set -e

    if [ ${EXIT_CODE} -eq 0 ]; then
        echo "Application exited successfully."
    else
        echo "Application exited with code ${EXIT_CODE}."
    fi

    echo "Now sleeping."
    sleep infinity
}

main "${@}"
