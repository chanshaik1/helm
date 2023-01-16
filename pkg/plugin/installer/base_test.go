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

package installer // import "helm.sh/helm/v3/pkg/plugin/installer"

import (
	"fmt"
	"os"
	"testing"
)

func TestPath(t *testing.T) {
	tests := []struct {
		source         string
		helmPluginsDir string
		expectPath     string
	}{
		{
			source:         "",
			helmPluginsDir: "/helm/data/plugins",
			expectPath:     "",
		}, {
			source:         "https://github.com/jkroepke/helm-secrets",
			helmPluginsDir: "/helm/data/plugins",
			expectPath:     "/helm/data/plugins/helm-secrets",
		}, {
			source:         "https://github.com/jkroepke/helm-secrets",
			helmPluginsDir: fmt.Sprintf("/helm/data/plugins%c/remote/helm/data/plugins", os.PathListSeparator),
			expectPath:     "/helm/data/plugins/helm-secrets",
		},
	}

	for _, tt := range tests {

		os.Setenv("HELM_PLUGINS", tt.helmPluginsDir)
		baseIns := newBase(tt.source)
		baseInsPath := baseIns.Path()
		if baseInsPath != tt.expectPath {
			t.Errorf("expected name %s, got %s", tt.expectPath, baseInsPath)
		}
		os.Unsetenv("HELM_PLUGINS")
	}
}
