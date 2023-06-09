package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/appscodepc/.kube/config", "location")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	fmt.Println(kubeconfig)
	if err != nil {
		fmt.Printf("error %s, creating config files", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error %s, creating in cluster config files", err.Error())

		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, creating clientset", err.Error())
	}
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container {
						{
							Name: "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name: "http",
									Protocol: apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	prompt()
	fmt.Println("Updating deployment...")

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := deploymentsClient.Get(context.Background(), "demo-deployment", metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		result.Spec.Replicas = int32Ptr(1)
		result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13"
		_, err = deploymentsClient.Update(context.Background(), result, metav1.UpdateOptions{})
		return err
	})

	if retryErr != nil {
		log.Fatal(err)
	}

	fmt.Println("Updated deployment...")
	prompt()
	fmt.Printf("Listing deployments in namespace %s: \n", apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, deploy := range list.Items {
		fmt.Printf("Deployment Name: %s and have %v replicas\n", deploy.Name, *deploy.Spec.Replicas)
	}

	prompt()

	fmt.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(context.Background(), "demo-deployment", metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted deployment...")


}
func int32Ptr(i int32) *int32 { return &i }
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}