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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authinternal "k8s.io/kubernetes/pkg/apis/authentication"
)

// TestTokenReviewTransformerForward tests the forward method of the
// TokenReviewTransformer.
func TestTokenReviewTransformerForward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     authinternal.TokenReview
		want   authinternal.TokenReview
	}{
		{
			name:   "test forward tokenreview",
			tenant: "111111",
			in: authinternal.TokenReview{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TokenReview",
					APIVersion: "authentication.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-tokenreview",
				},
				Spec: authinternal.TokenReviewSpec{
					Token: "014fbff9a07c...",
					Audiences: []string{
						"https://myserver.example.com",
						"https://myserver.internal.example.com",
					},
				},
			},
			want: authinternal.TokenReview{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TokenReview",
					APIVersion: "authentication.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-tokenreview",
				},
				Spec: authinternal.TokenReviewSpec{
					Token: "014fbff9a07c...",
					Audiences: []string{
						"https://myserver.example.com",
						"https://myserver.internal.example.com",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewTokenReviewTransformer()
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward tokenreview, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestTokenReviewTransformerBackward tests the backward method of the
// TokenReviewTransformer.
func TestTokenReviewTransformerBackward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     authinternal.TokenReview
		want   authinternal.TokenReview
	}{
		{
			name:   "test backward tokenreview",
			tenant: "111111",
			in: authinternal.TokenReview{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TokenReview",
					APIVersion: "authentication.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-tokenreview",
				},
				Status: authinternal.TokenReviewStatus{
					Authenticated: true,
					User: authinternal.UserInfo{
						Username: "system:serviceaccount:111111-my-ns:my-sa",
						UID:      "42",
						Groups: []string{
							"system:serviceaccounts:111111-my-ns",
						},
					},
					Audiences: []string{
						"https://myserver.example.com",
						"https://myserver.internal.example.com",
					},
				},
			},
			want: authinternal.TokenReview{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TokenReview",
					APIVersion: "authentication.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-tokenreview",
				},
				Status: authinternal.TokenReviewStatus{
					Authenticated: true,
					User: authinternal.UserInfo{
						Username: "system:serviceaccount:my-ns:my-sa",
						UID:      "42",
						Groups: []string{
							"system:serviceaccounts:my-ns",
						},
					},
					Audiences: []string{
						"https://myserver.example.com",
						"https://myserver.internal.example.com",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewTokenReviewTransformer()
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward tokenreview, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
