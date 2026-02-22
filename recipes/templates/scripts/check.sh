#!/bin/bash

#repo="monkeyx-net/shamogu"
repo="HarbourMasters/Shipwright"
version=$(curl -sf "https://api.github.com/repos/${repo}/releases/latest" | jq -r '.tag_name')

if [[ -n "$version" && "$version" != "null" ]]; then
    echo $version
    wget -O port.tar.gz "https://github.com/${repo}/archive/refs/tags/${version}.tar.gz"
else
    branch=$(curl -sf "https://api.github.com/repos/${repo}" | jq -r '.default_branch')
    wget -O port.zip "https://github.com/${repo}/archive/refs/heads/${branch}.zip"
fi
