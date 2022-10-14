/*
Copyright 2022 The KubeZoo Authors.

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
	quota "github.com/kubewharf/kubezoo/pkg/apis/quota"
	quotav1alpha1 "github.com/kubewharf/kubezoo/pkg/apis/quota/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

func init() {
	utilruntime.Must(apiextensionsv1beta1.AddToScheme(scheme.Scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(quotav1alpha1.AddToScheme(scheme.Scheme))
}

func NewV1CustomResourceDefinitions() []*apiextensionsv1.CustomResourceDefinition {
	crds := []*apiextensionsv1.CustomResourceDefinition{}
	crds = append(crds, quota.NewClusterResourceQuotaCRD())
	return crds
}
