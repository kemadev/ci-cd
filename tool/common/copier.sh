#!/usr/bin/env bash

set -euo pipefail

TEMPLATE_REPO="https://github.com/kemadev/repo-template"
CONFIG_PATH="config/copier/.copier-answers.yml"

USAGE="Usage: $(basename "${BASH_SOURCE[0]}") [command]
Commands:
  copy    Copy files from template repository (new repository initialization).
  update  Update files from template repository.
  -h, --help  Show this help message.
"

function check_dependencies() {
	if ! command -v copier &>/dev/null; then
		echo "Error: 'copier' command not found. Please install it first."
		echo "See: https://github.com/copier-org/copier"
		exit 1
	fi

	if ! command -v git &>/dev/null; then
		echo "Error: 'git' command not found. Please install it first."
		exit 1
	fi
}

function go_to_git_root() {
	local GIT_ROOT
	GIT_ROOT="$(git rev-parse --show-toplevel)"
	if [ -n "${GIT_ROOT}" ]; then
		cd "${GIT_ROOT}"
	else
		echo "Error: Not a git repository."
		exit 1
	fi
}

function main() {
	check_dependencies
	go_to_git_root

	local command="${1:-}"
	shift

	case "${command}" in
	copy)
		copier copy "${TEMPLATE_REPO}" .
		;;
	update)
		copier update --answers-file "${CONFIG_PATH}" .
		;;
	"-h" | "--help")
		echo "${USAGE}"
		;;
	*)
		echo "Unknown command: ${command}"
		echo "${USAGE}"
		exit 1
		;;
	esac
}

main "${@}"
