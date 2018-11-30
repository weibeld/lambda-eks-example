package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func HandleRequest() {
	createDeployment()
}

func main() {
	lambda.Start(HandleRequest)
}

func createDeployment() {

	// Path to kubeconfig file
	kubeconfig := "./kubeconfig"

	// Create a Config (k8s.io/client-go/rest)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create an API Clientset (k8s.io/client-go/kubernetes)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create an AppsV1Client (k8s.io/client-go/kubernetes/typed/apps/v1)
	appsV1Client := clientset.AppsV1()

	// Deployment name
	name := fmt.Sprint(uuid.New())

	// Specification of the Deployment (k8s.io/api/apps/v1)
	deploymentSpec := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(2); return &i }(),
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []coreV1.ContainerPort{
								{
									Name:          "http",
									Protocol:      coreV1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	deployment, err := appsV1Client.Deployments("default").Create(deploymentSpec)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %s\n", deployment.ObjectMeta.Name)

}
