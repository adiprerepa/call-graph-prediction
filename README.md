# Self-Optimizing Istio

The following is an experiment to see if it is possible to optimize request routing in Envoy & Istio. 

The optimization in question here is https://hotos23.hotcrp.com/doc/hotos23-paper226.pdf?cap=hcav226TEHRJMXkFqXPHFgdGgoCJDLC, in section 4.2 - Multi-Hop Cost Optimization. Dynamic multi-hop request routing requires two things:
1. Customizable host selection from the set of available upstream hosts.
2. Service call graph prediction based on request type.
	- Call graph prediction is useful because in most microservice deployments, requests of similar type (header, body, time, origin, etc) trigger similar call graphs.


## Deploying

```bash
kubectl apply -f slate.yaml
```

This will deploy a this WebAssembly plugin into the service mesh, and install the Stateful Session feature in Envoy.

It will also deploy the controller into the mesh.

The actual plugin is located in the root directory `wasm-plugin/main.go`.