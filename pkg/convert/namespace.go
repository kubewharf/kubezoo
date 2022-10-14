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
package convert

import (
	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	internal "k8s.io/kubernetes/pkg/apis/core"
)

type NamespaceTransformer struct {
}

var _ ObjectTransformer = &NamespaceTransformer{}

func NewNamespaceTransformer() *NamespaceTransformer {
	return &NamespaceTransformer{}
}

func (t *NamespaceTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	ns, ok := obj.(*internal.Namespace)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of namesapce")
	}
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	if v, ok := ns.Labels[common.TenantNamespaceLabelKey]; ok && v != tenantID {
		return nil, errors.Errorf("namespace label %s is protected by kubezoo, can not be modified", common.TenantNamespaceLabelKey)
	} else {
		ns.Labels[common.TenantNamespaceLabelKey] = tenantID
	}
	return ns, nil
}

func (t *NamespaceTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	return obj, nil
}
