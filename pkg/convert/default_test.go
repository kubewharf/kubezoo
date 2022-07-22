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
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TestConvertTenantObjectToUpstreamObject tests the
// ConvertTenantObjectToUpstreamObject methods of DefaultConvertor.
func TestConvertTenantObjectToUpstreamObject(t *testing.T) {
	tenant := "111111"
	originName := "good"
	originNamespace := "luck"
	originOwnerReferenceName := "myor"

	// Group 1: for namespaced resources.
	testCases1 := map[string]struct {
		pod                        v1.Pod
		hasOwnerReference          bool
		isOwnerReferenceNamespaced bool
	}{
		"Pod without owner reference": {
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      originName,
					Namespace: originNamespace,
				},
			},
			hasOwnerReference:          false,
			isOwnerReferenceNamespaced: false,
		},
		"Pod with namespaced owner reference": {
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      originName,
					Namespace: originNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "ReplicaSet",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: true,
		},
		"Pod with unnamespaced owner reference": {
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      originName,
					Namespace: originNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "PersistentVolume",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: false,
		},
	}

	for testName, testCase := range testCases1 {
		t.Run(testName, func(t *testing.T) {
			c := NewDefaultConvertor(NewOwnerReferenceTransformer(checkGroupKind))
			err := c.ConvertTenantObjectToUpstreamObject(&testCase.pod, tenant, true)
			if err != nil {
				t.Errorf("Failed ConvertTenantObjectToUpstreamObject with err %s", err)
			}

			upstreamNamespace := testCase.pod.GetNamespace()
			if upstreamNamespace != tenant+util.TenantIDSeparator+originNamespace {
				t.Errorf("Unexpected namespace.")
			}

			if testCase.hasOwnerReference {
				ownerReference := testCase.pod.GetOwnerReferences()
				if testCase.isOwnerReferenceNamespaced {
					if ownerReference[0].Name != originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				} else {
					if ownerReference[0].Name != tenant+util.TenantIDSeparator+originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				}
			}
		})
	}

	// Group 2: for unnamespaced resources.
	testCases2 := map[string]struct {
		pv                         v1.PersistentVolume
		hasOwnerReference          bool
		isOwnerReferenceNamespaced bool
	}{
		"PV without owner reference": {
			pv: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: originName,
				},
			},
			hasOwnerReference:          false,
			isOwnerReferenceNamespaced: false,
		},
		"PV with namespaced owner reference": {
			pv: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: originName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "ReplicaSet",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: true,
		},
		"PV with unnamespaced owner reference": {
			pv: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: originName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "PersistentVolume",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: false,
		},
	}

	for testName, testCase := range testCases2 {
		t.Run(testName, func(t *testing.T) {
			c := NewDefaultConvertor(NewOwnerReferenceTransformer(checkGroupKind))
			err := c.ConvertTenantObjectToUpstreamObject(&testCase.pv, tenant, false)
			if err != nil {
				t.Errorf("Failed ConvertTenantObjectToUpstreamObject with err %s", err)
			}

			upstreamName := testCase.pv.GetName()
			if upstreamName != tenant+util.TenantIDSeparator+originName {
				t.Errorf("Unexpected name.")
			}

			if testCase.hasOwnerReference {
				ownerReference := testCase.pv.GetOwnerReferences()
				if testCase.isOwnerReferenceNamespaced {
					if ownerReference[0].Name != originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				} else {
					if ownerReference[0].Name != tenant+util.TenantIDSeparator+originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				}
			}
		})
	}
}

// TestConvertUpstreamObjectToTenantObject tests the
// ConvertUpstreamObjectToTenantObject methods of DefaultConvertor.
func TestConvertUpstreamObjectToTenantObject(t *testing.T) {
	tenant := "111111"
	originName := tenant + util.TenantIDSeparator + "good"
	originNamespace := tenant + util.TenantIDSeparator + "luck"
	originOwnerReferenceName := tenant + util.TenantIDSeparator + "myor"

	// Group 1: for namespaced resources.
	testCases1 := map[string]struct {
		pod                        v1.Pod
		hasOwnerReference          bool
		isOwnerReferenceNamespaced bool
	}{
		"Pod without owner reference": {
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      originName,
					Namespace: originNamespace,
				},
			},
			hasOwnerReference:          false,
			isOwnerReferenceNamespaced: false,
		},
		"Pod with namespaced owner reference": {
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      originName,
					Namespace: originNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "ReplicaSet",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: true,
		},
		"Pod with unnamespaced owner reference": {
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      originName,
					Namespace: originNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "PersistentVolume",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: false,
		},
	}

	for testName, testCase := range testCases1 {
		t.Run(testName, func(t *testing.T) {
			c := NewDefaultConvertor(NewOwnerReferenceTransformer(checkGroupKind))
			err := c.ConvertUpstreamObjectToTenantObject(&testCase.pod, tenant, true)
			if err != nil {
				t.Errorf("Failed ConvertTenantObjectToUpstreamObject with err %s", err)
			}

			tenantNamespace := testCase.pod.GetNamespace()
			if tenant+util.TenantIDSeparator+tenantNamespace != originNamespace {
				t.Errorf("Unexpected namespace.")
			}

			if testCase.hasOwnerReference {
				ownerReference := testCase.pod.GetOwnerReferences()
				if testCase.isOwnerReferenceNamespaced {
					if ownerReference[0].Name != originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				} else {
					if tenant+util.TenantIDSeparator+ownerReference[0].Name != originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				}
			}
		})
	}

	// Group 2: for unnamespaced resources.
	testCases2 := map[string]struct {
		pv                         v1.PersistentVolume
		hasOwnerReference          bool
		isOwnerReferenceNamespaced bool
	}{
		"PV without owner reference": {
			pv: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: originName,
				},
			},
			hasOwnerReference:          false,
			isOwnerReferenceNamespaced: false,
		},
		"PV with namespaced owner reference": {
			pv: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: originName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "apps/v1",
							Kind:       "ReplicaSet",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: true,
		},
		"PV with unnamespaced owner reference": {
			pv: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: originName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "PersistentVolume",
							Name:       originOwnerReferenceName,
						},
					},
				},
			},
			hasOwnerReference:          true,
			isOwnerReferenceNamespaced: false,
		},
	}

	for testName, testCase := range testCases2 {
		t.Run(testName, func(t *testing.T) {
			c := NewDefaultConvertor(NewOwnerReferenceTransformer(checkGroupKind))
			err := c.ConvertUpstreamObjectToTenantObject(&testCase.pv, tenant, false)
			if err != nil {
				t.Errorf("Failed ConvertTenantObjectToUpstreamObject with err %s", err)
			}

			tenantName := testCase.pv.GetName()
			if tenant+util.TenantIDSeparator+tenantName != originName {
				t.Errorf("Unexpected name.")
			}

			if testCase.hasOwnerReference {
				ownerReference := testCase.pv.GetOwnerReferences()
				if testCase.isOwnerReferenceNamespaced {
					if ownerReference[0].Name != originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				} else {
					if tenant+util.TenantIDSeparator+ownerReference[0].Name != originOwnerReferenceName {
						t.Errorf("Unexpected owner reference name.")
					}
				}
			}
		})
	}
}
