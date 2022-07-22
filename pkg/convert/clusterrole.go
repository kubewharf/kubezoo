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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// ClusterRoleTransformer implements the transformation between
// client and upstream server for ClusterRole resource.
type ClusterRoleTransformer struct {
	listTenantCRDs ListTenantCRDsFunc
}

var _ ObjectTransformer = &ClusterRoleTransformer{}

// NewClusterRoleTransformer initiates a ClusterRoleTransformer
// which implements the ObjectTransformer interfaces.
func NewClusterRoleTransformer(listTenantCRDs ListTenantCRDsFunc) ObjectTransformer {
	return &ClusterRoleTransformer{
		listTenantCRDs: listTenantCRDs,
	}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *ClusterRoleTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	clusterRole, ok := obj.(*rbacinternal.ClusterRole)
	if !ok {
		return nil, errors.New("failed to assert runtime.Object to internal version of ClusterRole")
	}
	if err := t.transformRuleList(clusterRole.Rules, tenantID, transformRuleToUpstream); err != nil {
		return nil, errors.Wrap(err, "failed to transform rule list to upstream")
	}
	return clusterRole, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *ClusterRoleTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	clusterRole, ok := obj.(*rbacinternal.ClusterRole)
	if !ok {
		return nil, errors.New("failed to assert runtime.Object to internal version of ClusterRole")
	}
	if err := t.transformRuleList(clusterRole.Rules, tenantID, transformRuleToTenant); err != nil {
		return nil, errors.Wrap(err, "failed to transform rule list to tenant")
	}
	return clusterRole, nil
}

// transformRuleList transform the cluster role policy rules between upstream server
// and client side.
func (t *ClusterRoleTransformer) transformRuleList(ruleList []rbacinternal.PolicyRule, tenantID string, transform ruleTransformFunc) error {
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
