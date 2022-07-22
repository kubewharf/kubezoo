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
	crdinternal "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/util"
)

// CRDConvertor implements the transformation between client and
// upstream server for CustomResourceDefinition resource.
type CRDConvertor struct {
	ownerRefTransformer OwnerReferenceTransformer
}

var _ common.ObjectConvertor = &CRDConvertor{}

// NewCRDConvertor initiates a CRDConvertor which implements the
// ObjectConvertor interfaces.
func NewCRDConvertor(ort OwnerReferenceTransformer) common.ObjectConvertor {
	return &CRDConvertor{
		ownerRefTransformer: ort,
	}
}

// ConvertTenantObjectToUpstreamObject convert the tenant object to
// upstream object.
func (t *CRDConvertor) ConvertTenantObjectToUpstreamObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	crd, ok := obj.(*crdinternal.CustomResourceDefinition)
	if !ok {
		return errors.Errorf("fail to assert the runtime object to the internal version of crd")
	}

	if crd.Name != crd.Spec.Names.Plural+"."+crd.Spec.Group {
		return errors.Errorf("The CustomResourceDefinition \"%s\" is invalid: metadata.name: Invalid value: \"%s\": must be spec.names.plural+\".\"+spec.group",
			crd.Name, crd.Name)
	}
	crd.Spec.Group = util.AddTenantIDPrefix(tenantID, crd.Spec.Group)
	crd.Name = crd.Spec.Names.Plural + "." + crd.Spec.Group
	for i := range crd.OwnerReferences {
		target, err := t.ownerRefTransformer.Forward(&crd.OwnerReferences[i], tenantID)
		if err != nil {
			return err
		}
		crd.OwnerReferences[i] = *target
	}
	return nil
}

// ConvertUpstreamObjectToTenantObject convert the upstream object to
// tenant object.
func (t *CRDConvertor) ConvertUpstreamObjectToTenantObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	crd, ok := obj.(*crdinternal.CustomResourceDefinition)
	if !ok {
		return errors.Errorf("fail to assert the runtime object to the internal version of crd")
	}

	if !strings.HasPrefix(crd.Spec.Group, tenantID) {
		return errors.Errorf("invalid spec.group %s for crd %s, tenant id is %s", crd.Spec.Group, crd.Name, tenantID)
	}
	crd.Spec.Group = util.TrimTenantIDPrefix(tenantID, crd.Spec.Group)
	crd.Name = crd.Spec.Names.Plural + "." + crd.Spec.Group
	for i := range crd.OwnerReferences {
		target, err := t.ownerRefTransformer.Backward(&crd.OwnerReferences[i], tenantID)
		if err != nil {
			return err
		}
		crd.OwnerReferences[i] = *target
	}
	return nil
}
