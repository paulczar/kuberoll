/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	var kubeconfig *string
	var toDelete []string
	var finalPods, newPods []v1.Pod
	var runningOK bool
	//var found bool

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	namespace := flag.String("namespace", "default", "namespace containing deployment")
	label := flag.String("label", "", "labelSelector string (example: run=hello)")
	flag.Parse()

	if *label == "" {
		flag.Usage()
		os.Exit(1)
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("--> Listing pods with label %s in namespace %q\n", *label, *namespace)
	// Get intial list of pods with specified label
	for _, i := range listPods(clientset, *label, *namespace).Items {
		toDelete = append(toDelete, i.Name)
	}
	podCount := len(toDelete)
	fmt.Printf("    %s\n", strings.Join(toDelete[:], ", "))

	// Loop though pods and delete them one at a time
	for i := 0; i < podCount; i++ {
		var dp string
		dp, toDelete = toDelete[0], toDelete[1:]
		fmt.Printf("====> Delete %s.", dp)
		err = clientset.CoreV1().Pods(*namespace).Delete(dp, &metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		// Wait until pod is fully deleted.
		for {
			time.Sleep(2 * time.Second)
			fmt.Printf(".")
			_, err = clientset.CoreV1().Pods(*namespace).Get(dp, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				break
			}
		}
		fmt.Println()
		runningOK = false
		// loop through pod list until the new pod is created and running
		for !runningOK {
			time.Sleep(2 * time.Second)
			newList := listPods(clientset, *label, *namespace).Items
			newPods = []v1.Pod{}
			// Compare list to be deleted vs list of pods to get list of new pods
			for _, op := range toDelete {
				for _, np := range newList {
					if np.Name == op {
						continue
					} else {
						// If pod isn't in the original or final list of pods, we know its new.
						if !checkIfPodInList(np.Name, finalPods) {
							newPods = append(finalPods, np)
						}
						break
					}
				}
			}
			//fmt.Printf("====> New Pod(s) %s found", getPodNames(newPods))
			runningOK = true
			// check list of new pods, if running set true to end runningOK loop.
			for _, r := range newPods {
				if r.Status.Phase != v1.PodRunning {
					//fmt.Printf(".")
					runningOK = false
				} else {
					finalPods = append(finalPods, r)
				}
			}
			//fmt.Println()
		}
	}
}

func checkIfPodInList(name string, list []v1.Pod) bool {
	for _, p := range list {
		if name == p.Name {
			return true
		}
	}
	return false
}

func getPodNames(list []v1.Pod) string {
	var pods []string
	for _, d := range list {
		pods = append(pods, d.Name)
	}
	return strings.Join(pods[:], ",")
}

func listPods(clientset *kubernetes.Clientset, label, namespace string) *v1.PodList {
	var fs string
	list, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: label, FieldSelector: fs})
	if err != nil {
		panic(err)
	}
	return list
}
