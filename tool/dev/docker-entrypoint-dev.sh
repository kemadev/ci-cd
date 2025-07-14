#!/usr/bin/env sh

set -eu

main() {
    if [ -z "${GO_MAIN_FILE_DIR:-}" ]; then
        echo "GO_MAIN_FILE_DIR is not set."
        exit 1
    fi

    if [ -z "${HOME:-}" ]; then
        echo "HOME is not set."
        exit 1
    fi

    cp /run/secrets/netrc ~/.netrc

    BUILD_TARGET_PATH="${HOME}/app"

    go build -o "${BUILD_TARGET_PATH}" "${GO_MAIN_FILE_DIR}"

    set +e
    "${BUILD_TARGET_PATH}" "${@}"
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
