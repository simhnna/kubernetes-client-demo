/*
Copyright 2022.

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

package main

import (
	"context"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	nextaebcomv1alpha1 "my-domain/guestbook/api/v1alpha1"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(nextaebcomv1alpha1.AddToScheme(scheme))
}

func main() {
	home := homedir.HomeDir()
	// use the current context in kubeconfig
	config, _ := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &nextaebcomv1alpha1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		panic(err.Error())
	}
	platforms := nextaebcomv1alpha1.NextPlatformList{}
	err = client.Get().Resource("nextplatforms").Namespace("sia").Do(context.TODO()).Into(&platforms)
	if err != nil {
		panic(err)
	}
	for _, p := range platforms.Items {
		fmt.Printf("%s: %s\n", p.Name, p.Spec.PlatformVersion)
	}
	watcher, err := client.Get().Resource("nextplatforms").Namespace("sia").VersionedParams(&metav1.ListOptions{
		Watch: true,
		TypeMeta: metav1.TypeMeta{
			Kind:       "NextPlatform",
			APIVersion: "next.aeb.com/v1alpha1",
		},
	}, metav1.ParameterCodec).Watch(context.TODO())
	if err != nil {
		panic(err)
	}
	for o := range watcher.ResultChan() {
		fmt.Println(o.Type)
	}
}
