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
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// RoleTransformer implements the transformation between client and
// upstream server for Role resource.
type RoleTransformer struct {
	listTenantCRDs ListTenantCRDsFunc
}

var _ ObjectTransformer = &RoleTransformer{}

// NewRoleTransformer initiates a RoleTransformer which implements
// the ObjectTransformer interfaces.
func NewRoleTransformer(listTenantCRDs ListTenantCRDsFunc) ObjectTransformer {
	return &RoleTransformer{
		listTenantCRDs: listTenantCRDs,
	}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *RoleTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	role, ok := obj.(*rbacinternal.Role)
	if !ok {
		return nil, errors.Errorf("fail to assert object to internal version of role")
	}
	if err := t.transformRuleList(role.Rules, tenantID, transformRuleToUpstream); err != nil {
		return nil, errors.Wrap(err, "failed to transform rule list to upstream")
	}
	return role, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *RoleTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	role, ok := obj.(*rbacinternal.Role)
	if !ok {
		return nil, errors.Errorf("fail to assert object to internal version of role")
	}

	if err := t.transformRuleList(role.Rules, tenantID, transformRuleToTenant); err != nil {
		return nil, errors.Wrap(err, "failed to transform rule list to tenant")
	}
	return role, nil
}

// ListTenantCRDsFunc lists crds for tenant
type ListTenantCRDsFunc func(tenantID string) ([]*v1.CustomResourceDefinition, error)

// transformRuleList transform the role policy rules between upstream server
// and client side.
func (t *RoleTransformer) transformRuleList(ruleList []rbacinternal.PolicyRule, tenantID string, transform ruleTransformFunc) error {
	if len(ruleList) == 0 {
		return nil
	}
	crdList, err := t.listTenantCRDs(tenantID)
	if err != nil {
		return err
	}
	grm := util.NewCustomGroupResourcesMap(crdList)
	for i := range ruleList {
		if err := transform(&ruleList[i], grm, tenantID); err != nil {
			return err
		}
	}
	return nil
}

type ruleTransformFunc func(rule *rbacinternal.PolicyRule, grm util.CustomGroupResourcesMap, tenantID string) error

// transformRuleToUpstream transform the role policy rules from tenant
// object to upstream object.
func transformRuleToUpstream(rule *rbacinternal.PolicyRule, grm util.CustomGroupResourcesMap, tenantID string) error {
	for i := range rule.APIGroups {
		if grm.HasGroup(util.AddTenantIDPrefix(tenantID, rule.APIGroups[i])) {
			rule.APIGroups[i] = util.AddTenantIDPrefix(tenantID, rule.APIGroups[i])
		}
	}

	// todo: do we have a better way to handler resource names?
	resourceNamesToAdd := make([]string, 0)
	for _, resourceName := range rule.ResourceNames {
		resourceNamesToAdd = append(resourceNamesToAdd, util.AddTenantIDPrefix(tenantID, resourceName))
	}
	rule.ResourceNames = append(rule.ResourceNames, resourceNamesToAdd...)
	return nil
}

// transformRuleToTenant transform the role policy rules from upstream
// object to tenant object.
func transformRuleToTenant(rule *rbacinternal.PolicyRule, grm util.CustomGroupResourcesMap, tenantID string) error {
	for i := range rule.APIGroups {
		rule.APIGroups[i] = util.TrimTenantIDPrefix(tenantID, rule.APIGroups[i])
	}

	// todo: do we have a better way to handler resource names?
	filteredResourceNames := make([]string, 0)
	for _, resourceName := range rule.ResourceNames {
		if !strings.HasPrefix(resourceName, tenantID+"-") {
			filteredResourceNames = append(filteredResourceNames, resourceName)
		}
	}
	rule.ResourceNames = filteredResourceNames
	return nil
}
