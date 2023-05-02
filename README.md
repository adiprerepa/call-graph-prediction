# Self-Optimizing Istio

The following is an experiment to see if it is possible to optimize request routing in Envoy & Istio. 

## Deploying

```bash
kubectl apply -f config/slate-wasm-plugin.yaml
kubectl apply -f config/slate-envoyfilter.yaml
```

This will deploy a this WebAssembly plugin into the service mesh, and install the Stateful Session feature in Envoy.

The actual plugin is located in the root directory `main.go`.