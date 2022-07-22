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
	sa "k8s.io/apiserver/pkg/authentication/serviceaccount"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// RoleBindingTransformer implements the transformation between client
// and upstream server for RoleBinding resource.
type RoleBindingTransformer struct{}

var _ ObjectTransformer = &RoleBindingTransformer{}

// NewRoleBindingTransformer initiates a RoleBindingTransformer which
// implements the ObjectTransformer interfaces.
func NewRoleBindingTransformer() ObjectTransformer {
	return &RoleBindingTransformer{}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *RoleBindingTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	roleBinding, ok := obj.(*rbacinternal.RoleBinding)
	if !ok {
		return nil, errors.Errorf("fail to assert object to internal version of rolebinding")
	}

	if err := transformSubjectList(roleBinding.Subjects, tenantID, transformSubjectToUpstream); err != nil {
		return nil, errors.WithMessagef(err, "failed to transform subjects to upstream for roleBinding %s/%s", roleBinding.Namespace, roleBinding.Name)
	}
	if roleBinding.RoleRef.Kind == "ClusterRole" && len(roleBinding.RoleRef.Name) > 0 {
		roleBinding.RoleRef.Name = util.AddTenantIDPrefix(tenantID, roleBinding.RoleRef.Name)
	}
	return roleBinding, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *RoleBindingTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	roleBinding, ok := obj.(*rbacinternal.RoleBinding)
	if !ok {
		return nil, errors.Errorf("fail to assert object to internal version of rolebinding")
	}

	if err := transformSubjectList(roleBinding.Subjects, tenantID, transformSubjectToTenant); err != nil {
		return nil, errors.WithMessagef(err, "failed to transform subjects to tenant for roleBinding %s/%s", roleBinding.Namespace, roleBinding.Name)
	}
	if roleBinding.RoleRef.Kind == "ClusterRole" {
		if !strings.HasPrefix(roleBinding.RoleRef.Name, tenantID) {
			return nil, errors.Errorf("invalid roleRef name %s in roleBinding %s, tenant id is %s", roleBinding.RoleRef.Name, roleBinding.Name, tenantID)
		}
		roleBinding.RoleRef.Name = util.TrimTenantIDPrefix(tenantID, roleBinding.RoleRef.Name)
	}
	return roleBinding, nil
}

// transformRuleList transform the subjects between upstream server
// and client side.
func transformSubjectList(subjectList []rbacinternal.Subject, tenantID string, transform subjectTransformFunc) error {
	for i := range subjectList {
		if err := transform(&subjectList[i], tenantID); err != nil {
			return err
		}
	}
	return nil
}

type subjectTransformFunc func(subject *rbacinternal.Subject, tenantID string) error

// transformSubjectToUpstream transform the subject from tenant
// object to upstream object.
func transformSubjectToUpstream(subject *rbacinternal.Subject, tenantID string) error {
	switch subject.Kind {
	case rbacinternal.UserKind:
		if len(subject.Name) > 0 {
			subject.Name = util.AddTenantIDPrefix(tenantID, subject.Name)
		}
	case rbacinternal.ServiceAccountKind:
		if len(subject.Namespace) > 0 {
			subject.Namespace = util.AddTenantIDPrefix(tenantID, subject.Namespace)
		}
	case rbacinternal.GroupKind:
		if strings.HasPrefix(subject.Name, sa.ServiceAccountGroupPrefix) && len(subject.Name) > len(sa.ServiceAccountGroupPrefix) {
			namespace := strings.TrimPrefix(subject.Name, sa.ServiceAccountGroupPrefix)
			subject.Name = sa.MakeNamespaceGroupName(util.AddTenantIDPrefix(tenantID, namespace))
		}
	}
	return nil
}

// transformSubjectToTenant transform the subject from upstream
// object to tenant object.
func transformSubjectToTenant(subject *rbacinternal.Subject, tenantID string) error {
	switch subject.Kind {
	case rbacinternal.UserKind:
		if !strings.HasPrefix(subject.Name, tenantID) {
			return errors.Errorf("invalid subject name %s, should have tenant id %s as prefix", subject.Name, tenantID)
		}
		subject.Name = util.TrimTenantIDPrefix(tenantID, subject.Name)
	case rbacinternal.ServiceAccountKind:
		if !strings.HasPrefix(subject.Namespace, tenantID) {
			return errors.Errorf("invalid subject namespace %s, should have tenant id %s as prefix", subject.Namespace, tenantID)
		}
		subject.Namespace = util.TrimTenantIDPrefix(tenantID, subject.Namespace)
	case rbacinternal.GroupKind:
		if strings.HasPrefix(subject.Name, sa.ServiceAccountGroupPrefix) {
			namespace := strings.TrimPrefix(subject.Name, sa.ServiceAccountGroupPrefix)
			if !strings.HasPrefix(namespace, tenantID) {
				return errors.Errorf("invalid subject name %s, should have tenant id %s as prefix", subject.Name, tenantID)
			}
			subject.Name = sa.MakeNamespaceGroupName(util.TrimTenantIDPrefix(tenantID, namespace))
		}
	}
	return nil
}
