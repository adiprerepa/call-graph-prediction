package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
)

var ip string

func main() {
	port := ":8080"
	router := gin.Default()
	fmt.Printf("Starting server at port %s", port)
	router.GET("/getEndpoint", getEndpoint)
	router.Run(port)
}

func getEndpoint(c *gin.Context) {
	// get endpoints for a given service
	endpoints, err := GetEndpoints("httpbin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	selectedEndpoint := SelectEndpoint(endpoints)
	c.String(http.StatusOK, selectedEndpoint)
}

func SelectEndpoint(endpoints []string) string {
	// select an endpoint
	if len(endpoints) == 0 {
		return "no endpoints found"
	}
	return endpoints[0]
}

// for a given service, get the available endpoints
func GetEndpoints(serviceName string) ([]string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// get endpoints for a given service
	endpoints, err := clientset.CoreV1().Endpoints("default").List(context.TODO(),
		metav1.ListOptions{LabelSelector: "service=" + serviceName})
	if err != nil {
		return nil, err
	}
	// get the ip addresses of the endpoints
	var svcEndpoints []string
	for _, subset := range endpoints.Items[0].Subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				endpoint := fmt.Sprintf("%s:%d", address.IP, port.Port)
				fmt.Printf("got endpoint: %v\n", endpoint)
				svcEndpoints = append(svcEndpoints, endpoint)
			}
		}
	}
	return svcEndpoints, nil
}
