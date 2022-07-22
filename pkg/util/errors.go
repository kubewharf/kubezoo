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
	"errors"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
)

// TrimTenantIDFromError trims tenantID from error message and returns the new error.
func TrimTenantIDFromError(err error, tenantID string) error {
	switch t := err.(type) {
	case *apierrors.StatusError:
		return &apierrors.StatusError{
			ErrStatus: TrimTenantIDFromStatus(t.Status(), tenantID),
		}
	default:
		runtime.HandleError(fmt.Errorf("kubezoo received an error that is not an metav1.Status: %#+v", err))
		return errors.New(strings.ReplaceAll(err.Error(), tenantID+"-", ""))
	}
}

// TrimTenantIDFromStatus trims tenantID from status and returns the new status.
func TrimTenantIDFromStatus(status metav1.Status, tenantID string) metav1.Status {
	status.Message = strings.ReplaceAll(status.Message, tenantID+"-", "")
	if status.Details == nil {
		return status
	}
	status.Details.Name = strings.Replace(status.Details.Name, tenantID+"-", "", 1)
	status.Details.Group = TrimTenantIDPrefix(tenantID, status.Details.Group)
	for i := range status.Details.Causes {
		status.Details.Causes[i].Message = strings.ReplaceAll(status.Details.Causes[i].Message, tenantID+"-", "")
	}
	return status
}
