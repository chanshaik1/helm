/*
Copyright The Helm Authors.

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

package rules // import "helm.sh/helm/v3/pkg/lint/rules"

import "fmt"

// deprecatedApis lists APIs that are deprecated (left) with suggested alternatives (right).
//
// An empty rvalue indicates that the API is completely deprecated.
var deprecatedApis = map[string]string{
	"extensions/v1 Deployment":             "apps/v1 Deployment",
	"extensions/v1 DaemonSet":              "apps/v1 DaemonSet",
	"extensions/v1 ReplicaSet":             "apps/v1 ReplicaSet",
	"extensions/v1beta1 PodSecurityPolicy": "policy/v1beta1 PodSecurityPolicy",
	"extensions/v1beta1 NetworkPoliicy":    "networking.k8s.iio/v1beta1 NetworkPolicy",
	"apps/v1beta1 Deployment":              "apps/v1 Deployment",
	"apps/v1beta2 Deployment":              "apps/v1 Deployment",
}

// deprecatedAPIError indicates than an API is deprecated in Kubernetes
type deprecatedAPIError struct {
	Deprecated  string
	Alternative string
}

func (e deprecatedAPIError) Error() string {
	msg := fmt.Sprintf("the kind %q is deprecated", e.Deprecated)
	if e.Alternative != "" {
		msg += fmt.Sprintf(" in favor of %q", e.Alternative)
	}
	return msg
}

func validateNoDeprecations(resource *K8sYamlStruct) error {
	gvk := fmt.Sprintf("%s %s", resource.APIVersion, resource.Kind)
	if alt, ok := deprecatedApis[gvk]; ok {
		return deprecatedAPIError{
			Deprecated:  gvk,
			Alternative: alt,
		}
	}
	return nil
}
