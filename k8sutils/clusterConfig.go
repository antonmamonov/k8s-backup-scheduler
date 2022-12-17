// Copyright 2022 Anton Mamonov <hi@antonmamonov.com> GNU GENERAL PUBLIC LICENSE
package k8sutils

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClusterConfig struct {
	ClientSet *kubernetes.Clientset
}

func GetK8sClusterConfig() (*K8sClusterConfig, error) {

	// get the home directory path
	homeDir, homeDirError := os.UserHomeDir()

	if homeDirError != nil {
		return nil, homeDirError
	}

	fmt.Println("homeDir", homeDir)

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", homeDir+"/.kube/config")
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k8sClusterConfig := &K8sClusterConfig{
		ClientSet: clientset,
	}

	return k8sClusterConfig, nil
}

func CheckK8sClusterConfigConnection() (bool, error) {

	k8sClusterConfig, err := GetK8sClusterConfig()
	if err != nil {
		return false, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := k8sClusterConfig.ClientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	fmt.Printf("[CheckK8sClusterConfigConnection] There are %d pods in the cluster\n", len(pods.Items))

	// how can a cluster have 0 pods overall?
	if len(pods.Items) == 0 {
		return false, nil
	}

	return true, nil
}
