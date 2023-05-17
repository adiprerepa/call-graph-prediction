package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/types"
	"github.com/jaegertracing/jaeger/proto-gen/api_v3"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"time"
)

var (
	controller *KubeController
	requests   map[string]Request
)

type Request struct {
	TraceID string            `json:"traceId"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func init() {
	controller = NewKubeController()
	requests = make(map[string]Request)
}

func main() {
	// get the endpoint for the tracing grpc query service
	tracingEndpoint, err := controller.GetTracingQueryEndpoint()
	if err != nil {
		panic(err)
	}
	fmt.Printf("connecting to jaeger at %v\n", tracingEndpoint)
	conn, err := grpc.Dial(tracingEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := api_v3.NewQueryServiceClient(conn)
	traces, err := client.FindTraces(context.Background(), &api_v3.FindTracesRequest{
		Query: &api_v3.TraceQueryParameters{
			ServiceName: "istio-ingressgateway.istio-system",
			StartTimeMin: &types.Timestamp{
				Seconds: 0,
			},
			StartTimeMax: &types.Timestamp{
				Seconds: time.Now().Unix(),
			},
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		panic(err)
	}

	fmt.Printf("waiting for trace\n")
	trace, err := traces.Recv()
	if err != io.EOF {
		if err != nil {
			panic(err)
		}
		out, err := json.MarshalIndent(trace, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("trace: %v\n", string(out))
		//fmt.Printf("trace: %v\n", trace.String())
		children := make(map[string][]string)
		spanToService := make(map[string]string)
		graphRoot := ""

		for _, span := range trace.GetResourceSpans() {
			s := span.GetInstrumentationLibrarySpans()[0].GetSpans()[0]
			spanToService[string(s.SpanId)] = s.Name
			fmt.Printf("resource: at %v\n", s.Name)
			if len(s.ParentSpanId) == 0 {
				graphRoot = string(s.SpanId)
			} else {
				children[string(s.ParentSpanId)] = append(children[string(s.ParentSpanId)], string(s.SpanId))
			}
		}

		fmt.Printf("graphRoot: %v\n", spanToService[graphRoot])
		// print children
		for parent, children := range children {
			fmt.Printf("parent: %v, children: ", spanToService[parent])
			for _, child := range children {
				fmt.Printf("%v ", spanToService[child])
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("no trace found\n")
	}

	port := ":8080"
	router := gin.Default()
	fmt.Printf("Starting server at port %s", port)
	router.GET("/getEndpoint", getEndpoint)
	router.GET("/traces", getTraces)
	router.POST("/traces", addTrace)
	if err := router.Run(port); err != nil {
		panic(err)
	}
}

func getEndpoint(c *gin.Context) {
	// get endpoints for a given service
	endpoints, err := controller.GetEndpoints("httpbin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	selectedEndpoint := SelectEndpoint(endpoints)
	c.String(http.StatusOK, selectedEndpoint)
}

func getTraces(c *gin.Context) {
	// get traces from the requests map
	c.JSON(http.StatusOK, requests)
}

func addTrace(c *gin.Context) {
	// add a trace to the tracing service
	var req Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requests[req.TraceID] = req
	c.String(http.StatusOK, "trace added")
}

func SelectEndpoint(endpoints []string) string {
	// select an endpoint
	if len(endpoints) == 0 {
		return "no endpoints found"
	}
	return endpoints[0]
}
