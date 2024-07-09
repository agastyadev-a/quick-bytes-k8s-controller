package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"strings"
	"time"
)

func k8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func createToDoAppDeployment(installationName string, image string, namespace string, containerPort int32) *appsv1.Deployment {
	tm := time.Now()
	timeStr, _ := fmt.Printf("%s", tm)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "todo-deployment-" + installationName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "todoApp",
				"createdAt": strconv.Itoa(timeStr),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":              "todoApp",
					"installationName": installationName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":              "todoApp",
						"installationName": installationName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  installationName,
							Image: image,
							Env: []corev1.EnvVar{
								{
									Name:  "INSTALLATION_NAME",
									Value: installationName,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: *int32Ptr(containerPort),
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func int32Ptr(i int32) *int32 { return &i }

func deployTodoApp(clientset *kubernetes.Clientset, deployment *appsv1.Deployment) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := clientset.AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
		return err
	})
	if retryErr != nil {
		fmt.Printf("Failed to create deployment: %v\n", retryErr)
	}

	fmt.Println("Deployment created successfully!")
}

func deleteDeployment(deploymentsClient *kubernetes.Clientset, deployment *appsv1.Deployment) {
	fmt.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.AppsV1().Deployments(deployment.Namespace).Delete(context.TODO(), deployment.Name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	} else {
		fmt.Println("Deleted deployment...")
	}
}

func getDeployment(deploymentsClient *kubernetes.Clientset, namespace string) v1.Deployment {
	fmt.Println("Getting deployment...")

	deploymentSpec, err := deploymentsClient.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range deploymentSpec.Items {
		log.Log.Info(" name and replicas" + d.Name)
		if strings.HasPrefix(d.Name, "todo-deployment") {
			return d
		}

	}
	return v1.Deployment{}

}
