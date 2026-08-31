#!/usr/bin/env bash

hide_copied_tooling() {
    local implementation="$1"
    local source_root="$2"
    local copied_tooling="$3"
    if [[ "${implementation}" == shared && -d "${source_root}/.golib" ]]; then
        mv "${source_root}/.golib" "${copied_tooling}"
    fi
}

restore_copied_tooling() {
    local implementation="$1"
    local source_root="$2"
    local copied_tooling="$3"
    if [[ "${implementation}" == shared && -d "${copied_tooling}" &&
        ! -e "${source_root}/.golib" ]]; then
        mv "${copied_tooling}" "${source_root}/.golib"
    fi
}
