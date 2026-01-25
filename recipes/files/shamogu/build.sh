#!/bin/bash

echo ${{ matrix.arch }}
echo ${{ steps.parse-json.outputs.port_folder }}

#pwd
#ls -lha
#cd "new_ports/build/${PORT_FOLDER}"
#pwd
#ls -lha
#${PORT_BUILD}
#mkdir -p dist
#cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
#strip "dist/${PORT_EXE}.${ARCH}" || true
#tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
