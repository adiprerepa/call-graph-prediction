package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func getRoutingRules(c *gin.Context) {
	// get endpoints for a given service
	podName := c.GetHeader("x-slate-podname")
	var service string
	var err error
	if podName == "SLATE_UNKNOWN" {
		c.JSON(http.StatusOK, RouteMatchRules{})
		return
	} else {
		service, err = controller.GetPodServiceName(podName)
		if err != nil {
			fmt.Printf("could not get service name for pod %v\n", podName)
			c.JSON(http.StatusOK, RouteMatchRules{})
		}
	}
	fmt.Printf("getting routing rules for service %v\n", service)
	endpoints, err := controller.GetEndpoints(service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	selectedEndpoint := SelectEndpoint(endpoints)
	fooRules := RouteMatchRules{
		Rule: []Rule{
			{
				OverrideHost: selectedEndpoint,
				Match: Match{
					Headers: map[string]string{
						"aditya-the-goat": "true",
					},
				},
			},
		},
	}
	c.JSON(http.StatusOK, fooRules)
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
	podName := c.GetHeader("x-slate-podname")
	if podName == "SLATE_UNKNOWN" {
		c.String(http.StatusOK, "skipping adding trace, podname is SLATE_UNKNOWN")
		return
	} else {
		serviceName, err := controller.GetPodServiceName(podName)
		if err != nil {
			fmt.Printf("could not get service name for pod %v\n", podName)
			c.JSON(http.StatusOK, RouteMatchRules{})
			return
		}
		fmt.Printf("adding trace for service %v\n", serviceName)
		requests[req.TraceID][serviceName] = req
		c.String(http.StatusOK, "trace added")
	}
}
