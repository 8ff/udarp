export GOROOT="$(brew --prefix golang)/libexec"
cd cmd/wasm
cp $GOROOT/misc/wasm/wasm_exec.js ../../www/js/
GOOS=js GOARCH=wasm go build -o ../../www/assets/main.wasm