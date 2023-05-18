package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeController struct {
	clientset *kubernetes.Clientset
}

func NewKubeController() *KubeController {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	return &KubeController{
		clientset: clientset,
	}
}

func (k *KubeController) GetEndpoints(serviceName string) ([]string, error) {
	// get endpoints for a given service
	endpoints, err := k.clientset.CoreV1().Endpoints("default").List(context.TODO(),
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

// get the endpoint for the tracing grpc query service
func (k *KubeController) GetTracingQueryEndpoint() (string, error) {
	svc, err := k.clientset.CoreV1().Services("istio-system").Get(context.TODO(), "tracing", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return svc.Spec.ClusterIP + ":16685", nil
}

func (k *KubeController) GetPodServiceName(podName string) (string, error) {
	pod, err := k.clientset.CoreV1().Pods("default").Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return pod.Labels["app"], nil
}
