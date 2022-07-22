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
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// ClusterRoleBindingTransformer implements the transformation between
// client and upstream server for ClusterRoleBinding resource.
type ClusterRoleBindingTransformer struct{}

var _ ObjectTransformer = &ClusterRoleBindingTransformer{}

// NewClusterRoleBindingTransformer initiates a ClusterRoleBindingTransformer
// which implements the ObjectTransformer interfaces.
func NewClusterRoleBindingTransformer() ObjectTransformer {
	return &ClusterRoleBindingTransformer{}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *ClusterRoleBindingTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	crb, ok := obj.(*rbacinternal.ClusterRoleBinding)
	if !ok {
		return nil, errors.Errorf("fail to assert object to internal version of role")
	}

	if err := transformSubjectList(crb.Subjects, tenantID, transformSubjectToUpstream); err != nil {
		return nil, errors.WithMessagef(err, "failed to transform subjects to upstream for clusterRoleBinding %s", crb.Name)
	}
	if crb.RoleRef.Kind == "ClusterRole" && len(crb.RoleRef.Name) > 0 {
		crb.RoleRef.Name = util.AddTenantIDPrefix(tenantID, crb.RoleRef.Name)
	}
	return crb, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *ClusterRoleBindingTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	crb, ok := obj.(*rbacinternal.ClusterRoleBinding)
	if !ok {
		return nil, errors.Errorf("fail to assert object to internal version of role")
	}
	if err := transformSubjectList(crb.Subjects, tenantID, transformSubjectToTenant); err != nil {
		return nil, errors.WithMessagef(err, "failed to transform subjects to tenant for clusterRoleBinding %s", crb.Name)
	}
	if crb.RoleRef.Kind == "ClusterRole" {
		if !strings.HasPrefix(crb.RoleRef.Name, tenantID) {
			return nil, errors.Errorf("invalid roleRef name %s in clusterRoleBinding %s, tenant id is %s", crb.RoleRef.Name, crb.Name, tenantID)
		}
		crb.RoleRef.Name = util.TrimTenantIDPrefix(tenantID, crb.RoleRef.Name)
	}
	return crb, nil
}
