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

package test_rest

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	"github.com/kubewharf/kubezoo/pkg/apis/tenant/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/util"
)

// NewStrategy creates and returns a tenantStrategy instance
func NewStrategy(typer runtime.ObjectTyper) tenantStrategy {
	return tenantStrategy{typer, names.SimpleNameGenerator}
}

// GetAttrs returns labels.Set, fields.Set, and error in case the given runtime.Object is not a Flunder
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	apiserver, ok := obj.(*v1alpha1.Tenant)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Tenant")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), SelectableFields(apiserver), nil
}

// SelectableFields returns a field set that represents the object.
func SelectableFields(obj *v1alpha1.Tenant) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

// MatchFlunder is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchTenant(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

type tenantStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (tenantStrategy) NamespaceScoped() bool {
	return false
}

func (tenantStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (tenantStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (tenantStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	tenant := obj.(*v1alpha1.Tenant)
	err := util.ValidateTenantName(tenant.Name)
	if err != nil {
		return field.ErrorList{&field.Error{
			Type:     field.ErrorTypeInvalid,
			Field:    "name",
			BadValue: tenant.Name,
			Detail:   *err,
		}}
	}

	return nil
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (tenantStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string { return nil }

func (tenantStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (tenantStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (tenantStrategy) Canonicalize(obj runtime.Object) {
}

func (tenantStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

// WarningsOnUpdate returns warnings for the given update.
func (tenantStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
