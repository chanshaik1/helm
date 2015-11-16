/*
Copyright 2015 The Kubernetes Authors All rights reserved.
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

package manager

import (
	"time"
)

// This map defines the primitives that DM knows how to handle implicitly.
// TODO (iantw): Make these come from the resourcifier(?). Add more as appropriate...
var Primitives = map[string]bool{
	"Pod": true,
	"ReplicationController": true,
	"Service":               true,
	"Namespace":             true,
	"Volume":                true,
	"Endpoints":             true,
	"PersistentVolumeClaim": true,
	"PersistentVolume":      true,
}

// SchemaImport represents an import as declared in a schema file.
type SchemaImport struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// Schema is a partial DM schema. We only need access to the imports object at this level.
type Schema struct {
	Imports []SchemaImport `json:"imports"`
}

// Repository manages storage for all Deployment Manager entities, as well as
// the common operations to store, access and manage them.
type Repository interface {
	// Deployments.
	ListDeployments() ([]Deployment, error)
	GetDeployment(name string) (*Deployment, error)
	GetValidDeployment(name string) (*Deployment, error)
	CreateDeployment(name string) (*Deployment, error)
	DeleteDeployment(name string, forget bool) (*Deployment, error)

	// Manifests.
	AddManifest(deploymentName string, manifest *Manifest) error
	ListManifests(deploymentName string) (map[string]*Manifest, error)
	GetManifest(deploymentName string, manifestName string) (*Manifest, error)

	// Types.
	ListTypes() []string
	GetTypeInstances(typeName string) []*TypeInstance
	ClearTypeInstances(deploymentName string)
	SetTypeInstances(deploymentName string, instances map[string][]*TypeInstance)
}

// Deployment defines a deployment that describes
// the creation, modification and/or deletion of a set of resources.
type Deployment struct {
	Name       string               `json:"name"`
	ID         int                  `json:"id"`
	CreatedAt  time.Time            `json:"createdAt,omitempty"`
	DeployedAt time.Time            `json:"deployedAt,omitempty"`
	ModifiedAt time.Time            `json:"modifiedAt,omitempty"`
	DeletedAt  time.Time            `json:"deletedAt,omitempty"`
	Status     deploymentStatus     `json:"status,omitempty"`
	Current    *Configuration       `json:"current,omitEmpty"`
	Manifests  map[string]*Manifest `json:"manifests,omitempty"`
}

// NewDeployment creates a new deployment.
func NewDeployment(name string, id int) *Deployment {
	return &Deployment{Name: name, ID: id, CreatedAt: time.Now(), Status: CreatedStatus,
		Manifests: make(map[string]*Manifest, 0)}
}

// NewManifest creates a new manifest.
func NewManifest(deploymentName string, manifestName string) *Manifest {
	return &Manifest{Deployment: deploymentName, Name: manifestName}
}

// deploymentStatus is an enumeration type for the status of a deployment.
type deploymentStatus string

// These constants implement the deploymentStatus enumeration type.
const (
	CreatedStatus  deploymentStatus = "Created"
	DeletedStatus  deploymentStatus = "Deleted"
	DeployedStatus deploymentStatus = "Deployed"
	FailedStatus   deploymentStatus = "Failed"
	ModifiedStatus deploymentStatus = "Modified"
)

func (s deploymentStatus) String() string {
	return string(s)
}

// LayoutResource defines the structure of resources in the manifest layout.
type LayoutResource struct {
	Resource
	Layout
}

// Layout defines the structure of a layout as returned from expansion.
type Layout struct {
	Resources []*LayoutResource `json:"resources,omitempty"`
}

// Manifest contains the input configuration for a deployment, the fully
// expanded configuration, and the layout structure of the manifest.
//
type Manifest struct {
	Deployment     string         `json:"deployment"`
	Name           string         `json:"name"`
	InputConfig    *Template      `json:"inputConfig"`
	ExpandedConfig *Configuration `json:"expandedConfig,omitempty"`
	Layout         *Layout        `json:"layout,omitempty"`
}

// Template describes a set of resources to be deployed.
// Manager expands a Template into a Configuration, which
// describes the set in a form that can be instantiated.
type Template struct {
	Name    string        `json:"name"`
	Content string        `json:"content"`
	Imports []*ImportFile `json:"imports"`
}

// ImportFile describes a base64 encoded file imported by a Template.
type ImportFile struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content"`
}

// Configuration describes a set of resources in a form
// that can be instantiated.
type Configuration struct {
	Resources []*Resource `json:"resources"`
}

// Resource describes a resource in a configuration. A resource has
// a name, a type and a set of properties. The name and type are used
// to identify the resource in Kubernetes. The properties are passed
// to Kubernetes as the resource configuration.
type Resource struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TypeInstance defines the metadata for an instantiation of a template type
// in a deployment.
type TypeInstance struct {
	Name       string `json:"name"`       // instance name
	Type       string `json:"type"`       // instance type
	Deployment string `json:"deployment"` // deployment name
	Manifest   string `json:"manifest"`   // manifest name
	Path       string `json:"path"`       // JSON path within manifest
}
