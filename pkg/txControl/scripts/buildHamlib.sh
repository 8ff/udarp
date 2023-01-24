#!/bin/bash
VERSION="4.5.4"

# Fail function
fail() {
    echo "Error: $1"
    exit 1
}

# Get OS and architecture and make all lowercase
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')

# Verify that current directory is called "txControl" bin directoru exists, if not exit
CURRENT_DIRNAME=$(basename $(pwd))
if [ "${CURRENT_DIRNAME}" != "txControl" ]; then
    fail "Error: script must be ran from txControl directory"
    exit 1
fi

# Check if bin directory exists, if not create it
if [ ! -d "bin" ]; then
    mkdir bin || fail "Error: could not create bin directory"
fi



# Build hamlib
rm -rf ./tmpBuild && mkdir ./tmpBuild && cd ./tmpBuild
curl -L https://github.com/Hamlib/Hamlib/releases/download/${VERSION}/hamlib-${VERSION}.tar.gz -o hamlib.tar.gz || fail "Error: could not download hamlib"
tar xzvf hamlib.tar.gz || fail "Error: could not extract hamlib"
rm hamlib.tar.gz || fail "Error: could not remove hamlib.tar.gz"
cd hamlib-${VERSION} || fail "Error: could not cd into hamlib-${VERSION}"
./configure --disable-shared || fail "Error: could not configure hamlib"
make -j4 || fail "Error: could not make hamlib"
cp tests/rigctld ../../bin/rigctld-${OS}-${ARCH} || fail "Error: could not copy rigctld to bin directory"
cp tests/rigctl ../../bin/rigctl-${OS}-${ARCH} || fail "Error: could not copy rigctl to bin directory"
cd ../.. && rm -rf ./tmpBuild && echo "************** Done **************"