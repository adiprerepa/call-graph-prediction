package main

import (
	// b64 "encoding/base64"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	// Embed the default VM context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultVMContext
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
}

// Override types.DefaultPluginContext.
func (p *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	httpContext := &httpHeaders{contextID: contextID, pluginContext: p}
	queueId, err := proxywasm.RegisterSharedQueue("queue-" + string(contextID))
	if err != nil {
		proxywasm.LogCriticalf("error registering shared queue: %v", err)
		return httpContext
	}
	httpContext.bodyQueueId = queueId
	return httpContext
}

type httpHeaders struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	contextID     uint32
	pluginContext *pluginContext
	bodyQueueId	uint32
}

func (ctx *httpHeaders) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.LogCriticalf("OnHttpRequestHeaders %d", ctx.contextID)
	
	return types.ActionContinue
}

func (ctx *httpHeaders) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	proxywasm.LogCriticalf("OnHttpRequestBody %d", ctx.contextID)
	body, err := proxywasm.GetHttpRequestBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("error getting request body: %v", err)
	} else {
		proxywasm.LogCriticalf("request body: %s", string(body))
		// commit body to shared queue
		if err = proxywasm.EnqueueSharedQueue(ctx.bodyQueueId, body); err != nil {
			proxywasm.LogCriticalf("error enqueuing body: %v", err)
		} else {
			proxywasm.LogCriticalf("enqueued body")
		}
	}
	return types.ActionContinue
}

func (ctx *httpHeaders) OnHttpStreamDone() {
	proxywasm.LogCriticalf("OnHttpStreamDone %d", ctx.contextID)

	hdr, err := proxywasm.GetHttpRequestHeader("x-b3-sampled")
	if err != nil {
		proxywasm.LogCriticalf("error getting request header x-b3-sampled: %v", err)
		hdr = "0"
	}
	if hdr == "1" {
		// this request is sampled, we should send it to the controller
		headers, err := proxywasm.GetHttpRequestHeaders()
		if err != nil {
			proxywasm.LogCriticalf("error getting request headers: %v", err)
			return
		}
		body, err := proxywasm.DequeueSharedQueue(ctx.bodyQueueId)
		if err != nil {
			proxywasm.LogCriticalf("error dequeuing body: %v", err)
			// we'll just assume the body is empty
			body = []byte{}
		} else {
			proxywasm.LogCriticalf("dequeued body: %s", string(body))
		}
		traceId, err := proxywasm.GetHttpRequestHeader("x-b3-traceid")
		if err != nil {
			proxywasm.LogCriticalf("error getting request header x-b3-traceid: %v", err)
			return
		}
		// print traceId, headers, and body
		proxywasm.LogCriticalf("traceId: %s", traceId)
		proxywasm.LogCriticalf("headers: %v", headers)
		proxywasm.LogCriticalf("body: %s", string(body))
		// send to controller
		controllerHeaders := [][2]string{
			{":method", "POST"},
			{":authority", "slate-controller.default.svc.cluster.local"},
			{":path", "/addTrace"},
			{"content-type", "text/json"},
		}
		controllerBody := []byte(`{"traceId":"` + traceId + `","headers":` + string(encodeHeaders(headers)) + `,"body":"` + string(body) + `"}`)
		// print controller body
		proxywasm.LogCriticalf("controller body: %s", string(controllerBody))
		cuid, err := proxywasm.DispatchHttpCall("outbound|8000||slate-controller.default.svc.cluster.local", controllerHeaders, controllerBody, make([][2]string, 0), 5000, func(numHeaders int, bodySize int, numTrailers int) {
			proxywasm.LogCriticalf("http call response: %d %d %d", numHeaders, bodySize, numTrailers)
		})
		if err != nil {
			proxywasm.LogCriticalf("error dispatching http call: %v", err)
		} else {
			proxywasm.LogCriticalf("dispatched http call: %d", cuid)
		}
	} else {
		proxywasm.LogCriticalf("request not sampled")
	}
}

func makeTraceKey(ctxId uint32) string {
	return "trace-" + string(ctxId)
}

func makeHeaderKey(ctxId uint32) string {
	return "header-" + string(ctxId)
}

func makeBodyKey(ctxId uint32) string {
	return "body-" + string(ctxId)
}

// given a [][2]string of headers, json-encode them into a string without the json library
func encodeHeaders(headers [][2]string) []byte {
	var encoded string
	encoded += "{"
	for i, header := range headers {
		encoded += "\"" + header[0] + "\":\"" + header[1] + "\""
		if i < len(headers) - 1 {
			encoded += ","
		}
	}
	encoded += "}"
	return []byte(encoded)
}
