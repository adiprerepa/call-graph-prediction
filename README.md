# Self-Optimizing Istio

The following is an experiment to see if it is possible to optimize request routing in Envoy & Istio. 

## Deploying

```bash
kubectl apply -f slate.yaml
```

This will deploy a this WebAssembly plugin into the service mesh, and install the Stateful Session feature in Envoy.

It will also deploy the controller into the mesh.

The actual plugin is located in the root directory `wasm-plugin/main.go`.