package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"

	b64 "encoding/base64"
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
	return &httpHeaders{contextID: contextID, pluginContext: p}
}

type httpHeaders struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	contextID     uint32
	pluginContext *pluginContext
}

func (ctx *httpHeaders) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.LogCriticalf("OnHttpRequestHeaders %d", ctx.contextID)	
	controllerHeaders := [][2]string {
		{":method", "GET"},
		{":authority", "slate-controller.default.svc.cluster.local"},
		{":path", "/getEndpoint"},
		{"content-type", "text/plain"},
	}
	proxywasm.LogCriticalf("making http call to slate-controller.default.svc.cluster.local")
	proxywasm.DispatchHttpCall("outbound|8000||slate-controller.default.svc.cluster.local", controllerHeaders, nil, make([][2]string, 0), 5000, func(numHeaders int, bodySize int, numTrailers int) {
		responseBody, err := proxywasm.GetHttpCallResponseBody(0, bodySize)
		if err != nil {
			proxywasm.LogCriticalf("error getting response body: %v", err)
			return
		}
		proxywasm.LogCriticalf("http call response body: %s", string(responseBody))
		ip := string(responseBody)

		headerKey := "x-slate-session-header"
		containsSessionHeader := false
		outgoingRequestHeaders, err := proxywasm.GetHttpRequestHeaders()
		if err != nil {
			proxywasm.LogCriticalf("error getting request headers: %v", err)
		}
		for _, header := range outgoingRequestHeaders {
			if header[0] == headerKey {
				containsSessionHeader = true
			}
		}
		encoded := b64.URLEncoding.EncodeToString([]byte(ip))
		if !containsSessionHeader {
			proxywasm.LogCriticalf("header %v not found, adding value %s", headerKey, encoded)
			proxywasm.AddHttpRequestHeader(headerKey, encoded)
		} else {
			proxywasm.LogCriticalf("header %v found, replacing with value: %s", headerKey, encoded)
			proxywasm.ReplaceHttpRequestHeader(headerKey, encoded)
		}
	})

	
	return types.ActionContinue
}