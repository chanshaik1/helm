#!/usr/bin/env bash

# Copyright The Helm Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

mockgen -package discovery -destination internal/test/discovery/mock_discovery.go k8s.io/client-go/discovery CachedDiscoveryInterface
mockgen -package action -source pkg/action/action.go -destination internal/test/action/mock_action.go /action/action.go

# Add license information at the top of the generated files
