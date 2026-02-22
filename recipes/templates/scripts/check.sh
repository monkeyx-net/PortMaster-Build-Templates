#!/bin/bash

#repo="monkeyx-net/shamogu"
#repo="HarbourMasters/Shipwright"
#repo="https://codeberg.org/anaseto/harmonist/archive/v1.0.2.zip"
version=$(curl -sf "https://api.github.com/repos/${repo}/releases/latest" | jq -r '.tag_name')

if [[ -n "$version" && "$version" != "null" ]]; then
    echo $version
    wget -O port.tar.gz "https://github.com/${repo}/archive/refs/tags/${version}.tar.gz"
    port_checksum=$(sha256sum port.tar.gz | cut -d' ' -f1)
    echo "Checksum: $port_checksum"
elif [[ "$repo" == *"codeberg.org"* ]]; then
    wget -O port.zip ${repo}
    port_checksum=$(sha256sum port.zip | cut -d' ' -f1)
    echo "Checksum: $port_checksum"
else
    branch=$(curl -sf "https://api.github.com/repos/${repo}" | jq -r '.default_branch')
    wget -O port.zip "https://github.com/${repo}/archive/refs/heads/${branch}.zip"
    port_checksum=$(sha256sum port.zip | cut -d' ' -f1)
    echo "Checksum: $port_checksum"
fi
