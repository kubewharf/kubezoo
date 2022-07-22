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

package util

import (
	"fmt"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// TestTrimTenantIDFromErrorWithStatusError tests with status error.
func TestTrimTenantIDFromErrorWithStatusError(t *testing.T) {
	tenantId := "111111"
	msg := "unauthorized"
	statusErr := apierrors.NewUnauthorized(tenantId + "-" + msg)
	err := TrimTenantIDFromError(statusErr, tenantId)
	if err.Error() != msg {
		fmt.Printf("error message: %s", err)
		t.Errorf("unexpected trim tenant id from error")
	}
}

// TestTrimTenantIDFromErrorWithStatusError tests with non status error.
func TestTrimTenantIDFromErrorWithNonStatusError(t *testing.T) {
	tenantId := "111111"
	msg := "unauthorized"
	errIn := fmt.Errorf("%s-%s", tenantId, msg)
	err := TrimTenantIDFromError(errIn, tenantId)
	if err.Error() != msg {
		fmt.Printf("error message: %s", err)
		t.Errorf("unexpected trim tenant id from error")
	}
}
