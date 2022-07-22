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
	"reflect"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	internal "k8s.io/kubernetes/pkg/apis/core"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TestEventTransformerForward tests the forward method of the
// EventTransformer.
func TestEventTransformerForward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     internal.Event
		want   internal.Event
	}{
		{
			name:   "test forward events of namespaced object",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Namespace:  "default",
					Name:       "pod-2",
				},
				Reason:  "Scheduled",
				Message: "Successfully assigned default/pod-2 to node-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Namespace:  "111111-default",
					Name:       "pod-2",
				},
				Reason:  "Scheduled",
				Message: "Successfully assigned default/pod-2 to node-1",
				Type:    "Normal",
			},
		},
		{
			name:   "test forward events of nonNamespaced object",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "PersistentVolume",
					Name:       "pv-1",
				},
				Reason:  "Created",
				Message: "Successfully created pv-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "PersistentVolume",
					Name:       "111111-pv-1",
				},
				Reason:  "Created",
				Message: "Successfully created pv-1",
				Type:    "Normal",
			},
		},
		{
			name:   "test forward events of namespaced custom resource",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "kubezoo.io/v1",
					Kind:       "Foo",
					Name:       "foo-1",
					Namespace:  "default",
				},
				Reason:  "Created",
				Message: "Successfully created default/foo-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "111111-kubezoo.io/v1",
					Kind:       "Foo",
					Name:       "foo-1",
					Namespace:  "111111-default",
				},
				Reason:  "Created",
				Message: "Successfully created default/foo-1",
				Type:    "Normal",
			},
		},
		{
			name:   "test forward events of nonNamespaced custom resource",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "kubezoo.io/v1",
					Kind:       "Bar",
					Name:       "bar-1",
					Namespace:  "default",
				},
				Reason:  "Created",
				Message: "Successfully created default/bar-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "111111-kubezoo.io/v1",
					Kind:       "Bar",
					Name:       "111111-bar-1",
					Namespace:  "default",
				},
				Reason:  "Created",
				Message: "Successfully created default/bar-1",
				Type:    "Normal",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewEventTransformer(NewObjectReferenceTransformer(checkGroupKindForAllResources))
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward event, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestEventTransformerBackward tests the backward method of the
// EventTransformer.
func TestEventTransformerBackward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     internal.Event
		want   internal.Event
	}{
		{
			name:   "test backward events of namespaced object",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Namespace:  "111111-default",
					Name:       "pod-2",
				},
				Reason:  "Scheduled",
				Message: "Successfully assigned default/pod-2 to node-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Namespace:  "default",
					Name:       "pod-2",
				},
				Reason:  "Scheduled",
				Message: "Successfully assigned default/pod-2 to node-1",
				Type:    "Normal",
			},
		},
		{
			name:   "test backward events of nonNamespaced object",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "PersistentVolume",
					Name:       "111111-pv-1",
				},
				Reason:  "Created",
				Message: "Successfully created pv-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "v1",
					Kind:       "PersistentVolume",
					Name:       "pv-1",
				},
				Reason:  "Created",
				Message: "Successfully created pv-1",
				Type:    "Normal",
			},
		},
		{
			name:   "test backward events of namespaced custom resource",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "111111-kubezoo.io/v1",
					Kind:       "Foo",
					Name:       "foo-1",
					Namespace:  "111111-default",
				},
				Reason:  "Created",
				Message: "Successfully created default/foo-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "kubezoo.io/v1",
					Kind:       "Foo",
					Name:       "foo-1",
					Namespace:  "default",
				},
				Reason:  "Created",
				Message: "Successfully created default/foo-1",
				Type:    "Normal",
			},
		},
		{
			name:   "test backward events of nonNamespaced custom resource",
			tenant: "111111",
			in: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "111111-kubezoo.io/v1",
					Kind:       "Bar",
					Name:       "111111-bar-1",
					Namespace:  "default",
				},
				Reason:  "Created",
				Message: "Successfully created default/bar-1",
				Type:    "Normal",
			},
			want: internal.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-event",
				},
				InvolvedObject: internal.ObjectReference{
					APIVersion: "kubezoo.io/v1",
					Kind:       "Bar",
					Name:       "bar-1",
					Namespace:  "default",
				},
				Reason:  "Created",
				Message: "Successfully created default/bar-1",
				Type:    "Normal",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewEventTransformer(NewObjectReferenceTransformer(checkGroupKindForAllResources))
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward event, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// checkGroupKindFor checks whether NATIVE/CRD group/kind is namespaced
// It is only used for build unit tests
func checkGroupKindForAllResources(group, kind, tenantID string, isTenantObject bool) (namespaced, customResourceGroup bool, err error) {
	namespaced, err = util.IsGroupKindNamespaced(metav1.GroupKind{Group: group, Kind: kind})
	if err == nil {
		return namespaced, false, nil
	}

	fakeListTenantCRDsFunc := func(tenantID string) ([]*apiextensionsv1.CustomResourceDefinition, error) {
		return []*apiextensionsv1.CustomResourceDefinition{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foos.111111-kubezoo.io",
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Group: "111111-kubezoo.io",
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name: "v1",
						},
					},
					Scope: apiextensionsv1.NamespaceScoped,
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Plural: "foos",
						Kind:   "Foo",
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "bars.111111-kubezoo.io",
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Group: "111111-kubezoo.io",
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name: "v1",
						},
					},
					Scope: apiextensionsv1.ClusterScoped,
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Plural: "bars",
						Kind:   "Bar",
					},
				},
			},
		}, nil
	}

	crdList, err := fakeListTenantCRDsFunc(tenantID)
	if err != nil {
		return
	}
	// tenant crd group/kind
	if isTenantObject {
		group = util.AddTenantIDPrefix(tenantID, group)
	}
	for _, crd := range crdList {
		if crd.Spec.Group == group && crd.Spec.Names.Kind == kind {
			return crd.Spec.Scope == apiextensionsv1.NamespaceScoped, true, nil
		}
	}

	return
}
