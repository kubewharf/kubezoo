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

package proxy

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"

	"github.com/kubewharf/kubezoo/pkg/util"
)

type emptyWatch struct {
	resultChan chan watch.Event
}

// Stop implements Interface
func (ew *emptyWatch) Stop() {}

// ResultChan implements Interface
func (ew *emptyWatch) ResultChan() <-chan watch.Event {
	return ew.resultChan
}

// proxyWatch implements the proxy for watch method, including
// the converting between tenant object and upstream object.
type proxyWatch struct {
	tenantProxy *tenantProxy
	tenantID    string
	watch       watch.Interface
	stopChan    chan struct{}
	resultChan  chan watch.Event
}

// newProxyWatch return a watch proxy.
func newProxyWatch(w watch.Interface, tenantProxy *tenantProxy, tenantID string) (*proxyWatch, error) {
	pw := &proxyWatch{
		watch:       w,
		tenantProxy: tenantProxy,
		stopChan:    make(chan struct{}),
		resultChan:  make(chan watch.Event, 0),
		tenantID:    tenantID,
	}

	go pw.proxy()
	return pw, nil
}

// proxy is the main approach of receive and send the events
// between upstream server and user client.
func (w *proxyWatch) proxy() {
	defer close(w.resultChan)

	klog.V(4).Info("Starting proxyWatcher.")
	defer klog.V(4).Info("Stopping proxyWatcher.")

	ch := w.watch.ResultChan()
	defer w.watch.Stop()
	for {
		select {
		case <-w.stopChan:
			return

		case event, ok := <-ch:
			if !ok {
				// End of results
				return
			}

			switch event.Type {
			case watch.Added, watch.Modified, watch.Deleted:
				obj := event.Object
				if !util.UpstreamObjectBelongsToTenant(obj, w.tenantID, w.tenantProxy.namespaceScoped) {
					continue
				}

				utd := obj.(*unstructured.Unstructured)
				output := w.tenantProxy.New()
				if err := w.tenantProxy.convertUnstructuredToOutput(utd, output); err != nil {
					err := fmt.Errorf("failed to convert unstructured to output, utd: %+v, err: %v", utd, err)
					klog.Error(err)
					w.send(newInternalErrorEvent(err))
					return
				}

				if err := w.tenantProxy.convertUpstreamObjectToTenantObject(output, w.tenantID); err != nil {
					err := fmt.Errorf("failed to convert upstream object to tenant object, utd: %+v, err: %v", utd, err)
					klog.Error(err)
					w.send(newInternalErrorEvent(err))
					return
				}

				event.Object = output
				w.send(event)

			case watch.Bookmark:
				w.send(event)

			case watch.Error:
				errObject := errors.FromObject(event.Object)
				statusErr, ok := errObject.(*errors.StatusError)
				if !ok {
					klog.Error(spew.Sprintf("Received an error which is not *metav1.Status but %#+v", event.Object))
					return
				}
				status := util.TrimTenantIDFromStatus(statusErr.Status(), w.tenantID)
				event.Object = &status
				w.send(event)
				return

			default:
				err := fmt.Errorf("failed to recognize Event type %q", event.Type)
				klog.Error(err)
				w.send(newInternalErrorEvent(err))
				return
			}
		}
	}
}

// newInternalErrorEvent create a error event.
func newInternalErrorEvent(err error) watch.Event {
	return watch.Event{
		Type:   watch.Error,
		Object: &errors.NewInternalError(err).ErrStatus,
	}
}

// send sends the event to resultChan.
func (w *proxyWatch) send(event watch.Event) {
	// Writing to an unbuffered channel is blocking operation
	// and we need to check if stop wasn't requested while doing so.
	select {
	case w.resultChan <- event:
	case <-w.stopChan:
	}
}

// Stop implements Interface
func (w *proxyWatch) Stop() {
	close(w.stopChan)
}

// ResultChan implements Interface
func (w *proxyWatch) ResultChan() <-chan watch.Event {
	return w.resultChan
}
