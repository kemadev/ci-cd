#!/usr/bin/env sh

set -eu

main() {
    if [ -z "${GO_MAIN_FILE_DIR:-}" ]; then
        echo "GO_MAIN_FILE_DIR is not set."
        exit 1
    fi

    cp /run/secrets/netrc ~/.netrc

    go build -o /app "${GO_MAIN_FILE_DIR}"

    /app "${@}"

    echo "Failure running the application, now sleeping."
    sleep infinity
    exit 1
}

main "${@}"
