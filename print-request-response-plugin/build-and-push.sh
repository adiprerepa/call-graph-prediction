tinygo build -o wasm-out/print_slate_plugin.wasm -scheduler=none -target=wasi main.go
docker build -t ghcr.io/adiprerepa/print-slate-plugin:latest .
docker push ghcr.io/adiprerepa/print-slate-plugin:latest
