/*
Copyright 2018 The Kubernetes Authors.

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

package validating

import (
	"context"
	"fmt"
	"net/http"

	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	providerv1 "sigs.k8s.io/cluster-api-provider-openstack/pkg/apis/openstackproviderconfig/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

func init() {
	webhookName := "validating-create-update-machine"
	if HandlerMap[webhookName] == nil {
		HandlerMap[webhookName] = []admission.Handler{}
	}
	HandlerMap[webhookName] = append(HandlerMap[webhookName], &CreateUpdateHandler{})
}

// CreateUpdateHandler handles Machine
type CreateUpdateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - import sigs.k8s.io/controller-runtime/pkg/client
	// - uncomment the InjectClient method at the bottom of this file.
	// Client  client.Client

	// Decoder decodes objects
	Decoder types.Decoder
}

func (h *CreateUpdateHandler) validatingFn(ctx context.Context, obj *clusterv1.Machine) (bool, string, error) {
	if obj.Spec.ProviderSpec.Value == nil {
		return false, "empty providerSpec is not allowed", nil
	}
	coreObj := metav1.TypeMeta{}
	err := yaml.Unmarshal(obj.Spec.ProviderSpec.Value.Raw, &coreObj)
	if err != nil {
		return false, "", fmt.Errorf("unable to parse providerSpec into metav1.TypeMeta: %v", err)
	}
	kind := coreObj.GetObjectKind()
	gv := kind.GroupVersionKind()
	if gv.Group != "openstackproviderconfig" && gv.Kind != "OpenstackProviderSpec" {
		return true, "allowed to be admitted, ignoring", nil
	}

	m, err := providerv1.MachineSpecFromProviderSpec(obj.Spec.ProviderSpec)
	if err != nil {
		return false, "", fmt.Errorf("unable to unmarshal providerSpec: %v", err)
	}

	return m.Validate()
}

var _ admission.Handler = &CreateUpdateHandler{}

// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req types.Request) types.Response {
	obj := &clusterv1.Machine{}

	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	allowed, reason, err := h.validatingFn(ctx, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	return admission.ValidationResponse(allowed, reason)
}

//var _ inject.Client = &CreateUpdateHandler{}
//
//// InjectClient injects the client into the CreateUpdateHandler
//func (h *CreateUpdateHandler) InjectClient(c client.Client) error {
//	h.Client = c
//	return nil
//}

var _ inject.Decoder = &CreateUpdateHandler{}

// InjectDecoder injects the decoder into the CreateUpdateHandler
func (h *CreateUpdateHandler) InjectDecoder(d types.Decoder) error {
	h.Decoder = d
	return nil
}
