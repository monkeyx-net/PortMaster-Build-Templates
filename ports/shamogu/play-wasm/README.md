To make the wasm-based version, you have to compile the package to wasm
into a `shamogu.wasm` file located in the same directory, along `index.html` and
`style.css`.

    GOOS=js GOARCH=wasm go build -o shamogu.wasm

You also need to copy the `wasm_exec.js` support file corresponding to your Go
version, as explained in [Go Wiki:
WebAssembly](https://go.dev/wiki/WebAssembly).

The resulting directory can then be served over HTTP.
