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
	sa "k8s.io/apiserver/pkg/authentication/serviceaccount"
	authinternal "k8s.io/kubernetes/pkg/apis/authentication"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TokenReviewTransformer implements the transformation between client
// and upstream server for TokenReview resource.
type TokenReviewTransformer struct{}

var _ ObjectTransformer = &TokenReviewTransformer{}

// NewTokenReviewTransformer initiates a TokenReviewTransformer which
// implements the ObjectTransformer interfaces.
func NewTokenReviewTransformer() ObjectTransformer {
	return &TokenReviewTransformer{}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *TokenReviewTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	return obj, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *TokenReviewTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	tokenReview, ok := obj.(*authinternal.TokenReview)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of tokenreview")
	}

	prefixedNamespace, name, err := sa.SplitUsername(tokenReview.Status.User.Username)
	if err != nil {
		// not a service account token
		return tokenReview, nil
	}
	namespace := util.TrimTenantIDPrefix(tenantID, prefixedNamespace)
	tokenReview.Status.User.Username = sa.MakeUsername(namespace, name)

	prefixedGroup := sa.MakeNamespaceGroupName(prefixedNamespace)
	for i := range tokenReview.Status.User.Groups {
		if tokenReview.Status.User.Groups[i] == prefixedGroup {
			tokenReview.Status.User.Groups[i] = sa.MakeNamespaceGroupName(namespace)
		}
	}
	return tokenReview, nil
}
