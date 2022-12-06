// Copyright 2022 The KubeZoo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DefaultRetry is the recommended retry for a conflict where multiple clients
// are making changes to the same resource.
var DefaultRetry = wait.Backoff{
	Steps:    5,
	Duration: 10 * time.Millisecond,
	Factor:   1.0,
	Jitter:   0.1,
}

type Updater interface {
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
}

func UpdateOnConflict(
	ctx context.Context,
	backoff wait.Backoff,
	apiReader client.Reader,
	apiWriter Updater,
	obj client.Object,
	f controllerutil.MutateFn,
) (controllerutil.OperationResult, error) {
	key := client.ObjectKeyFromObject(obj)
	firstTry := true
	result := controllerutil.OperationResultNone
	err := retry.RetryOnConflict(
		backoff,
		func() error {
			if firstTry {
				firstTry = false
			} else {
				// refresh obj from apiserver
				if err := apiReader.Get(ctx, key, obj); err != nil {
					return err
				}
			}
			existing := obj.DeepCopyObject()
			if err := mutate(f, key, obj); err != nil {
				return err
			}
			if equality.Semantic.DeepEqual(existing, obj) {
				// unchanged, finish trying
				return nil
			}
			if err := apiWriter.Update(ctx, obj); err != nil {
				// maybe conflict
				return err
			}
			// updated
			result = controllerutil.OperationResultUpdated
			return nil
		},
	)
	if err != nil {
		return controllerutil.OperationResultNone, err
	}
	return result, nil
}

// mutate wraps a MutateFn and applies validation to its result
func mutate(f controllerutil.MutateFn, key client.ObjectKey, obj client.Object) error {
	if err := f(); err != nil {
		return err
	}
	if newKey := client.ObjectKeyFromObject(obj); key != newKey {
		return fmt.Errorf("mutateFn cannot mutate object name and/or object namespace")
	}
	return nil
}
