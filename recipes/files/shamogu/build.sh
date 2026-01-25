#!/bin/bash
PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"



echo ${PORT_FOLDER}
exit 1

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
