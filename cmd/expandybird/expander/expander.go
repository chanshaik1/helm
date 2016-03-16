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

package expander

import (
	"bytes"
	"fmt"
	"encoding/json"
	"log"
	"os/exec"

	"github.com/ghodss/yaml"
	"github.com/kubernetes/helm/pkg/common"
)

// Expander abstracts interactions with the expander and deployer services.
type Expander interface {
	ExpandTemplate(template *common.Template) (string, error)
}

type expander struct {
	ExpansionBinary string
}

// NewExpander returns a new initialized Expander.
func NewExpander(binary string) Expander {
	return &expander{binary}
}

// ExpansionResult describes the unmarshalled output of ExpandTemplate.
type ExpansionResult struct {
	Config map[string]interface{}
	Layout map[string]interface{}
}

// NewExpansionResult creates and returns a new expansion result from
// the raw output of ExpandTemplate.
func NewExpansionResult(output string) (*ExpansionResult, error) {
	eResponse := &ExpansionResult{}
	if err := yaml.Unmarshal([]byte(output), eResponse); err != nil {
		return nil, fmt.Errorf("cannot unmarshal expansion result (%s):\n%s", err, output)
	}

	return eResponse, nil
}

// Marshal creates and returns an ExpansionResponse from an ExpansionResult.
func (eResult *ExpansionResult) Marshal() (*ExpansionResponse, error) {
	configYaml, err := yaml.Marshal(eResult.Config)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal manifest template (%s):\n%s", err, eResult.Config)
	}

	layoutYaml, err := yaml.Marshal(eResult.Layout)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal manifest layout (%s):\n%s", err, eResult.Layout)
	}

	return &ExpansionResponse{
		Config: string(configYaml),
		Layout: string(layoutYaml),
	}, nil
}

// ExpansionResponse describes the results of marshaling an ExpansionResult.
type ExpansionResponse struct {
	Config string `json:"config"`
	Layout string `json:"layout"`
}

// NewExpansionResponse creates and returns a new expansion response from
// the raw output of ExpandTemplate.
func NewExpansionResponse(output string) (*ExpansionResponse, error) {
	eResult, err := NewExpansionResult(output)
	if err != nil {
		return nil, err
	}

	eResponse, err := eResult.Marshal()
	if err != nil {
		return nil, err
	}

	return eResponse, nil
}

// Unmarshal creates and returns an ExpansionResult from an ExpansionResponse.
func (eResponse *ExpansionResponse) Unmarshal() (*ExpansionResult, error) {
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(eResponse.Config), &config); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config (%s):\n%s", err, eResponse.Config)
	}

	var layout map[string]interface{}
	if err := yaml.Unmarshal([]byte(eResponse.Layout), &layout); err != nil {
		return nil, fmt.Errorf("cannot unmarshal layout (%s):\n%s", err, eResponse.Layout)
	}

	return &ExpansionResult{
		Config: config,
		Layout: layout,
	}, nil
}

// ExpandTemplate passes the given configuration to the expander and returns the
// expanded configuration as a string on success.
func (e *expander) ExpandTemplate(template *common.Template) (string, error) {
	chartInv := template.ChartInvocation
	chart := template.Chart

	if chart.Expander.Name != "Expandybird" {
		message := fmt.Sprintf("Expandybird cannot do this kind of expansion: ", chart.Expander)
		return "", fmt.Errorf("error expanding template %s: %s", chartInv.Name, message)
	}

	if e.ExpansionBinary == "" {
		message := fmt.Sprintf("expansion binary cannot be empty")
		return "", fmt.Errorf("error expanding chart %s: %s", chartInv.Name, message)
	}

	if len(chart.Files) != 1 {
		message := fmt.Sprintf("Expandybird can only process a chart containing 1 file, got %d", len(chart.Files))
		return "", fmt.Errorf("error expanding template %s: %s", chartInv.Name, message)
	}

	entrypoint_file := chart.Files[0]

	if entrypoint_file.Path != chart.Expander.Entrypoint {
		message := fmt.Sprintf("The entrypoint in the chart.yaml canont be found %s", chart.Expander.Entrypoint)
		return "", fmt.Errorf("error expanding template %s: %s", chartInv.Name, message)
	}

	// Those are automatically increasing buffers, so writing arbitrary large
	// data here won't block the child process.
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Now we convert the new chart representation into the form that classic Expandybird takes.

	chartInvJson, err := json.Marshal(chartInv)
	if err != nil {
		return "", fmt.Errorf("error marshalling chart invocation %s: %s", chartInv.Name, err)
	}
	content := "{ \"resources\": [" + string(chartInvJson) + "] }"

	cmd := &exec.Cmd{
		Path: e.ExpansionBinary,
		// Note, that binary name still has to be passed argv[0].
		Args: []string{e.ExpansionBinary, content},
		// TODO(vagababov): figure out whether do we even need "PROJECT" and
		// "DEPLOYMENT_NAME" variables here.
		// [dcunnin]: Assuming we won't need this, at least for now.
		// Env:    append(os.Environ(), "PROJECT="+chartInv.Name, "DEPLOYMENT_NAME="+chartInv.Name),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	// Add entrypoint file (currently only 1 is supported).
	cmd.Args = append(cmd.Args, chart.Name, entrypoint_file.Path, entrypoint_file.Content)

	// Add schema file
	schema_name := chart.Name + ".schema"
	schemaJson, err := json.Marshal(chart.Schema)
	if err != nil {
		return "", fmt.Errorf("error marshalling schema %s: %s", chartInv.Name, err)
	}
	cmd.Args = append(cmd.Args, schema_name, schema_name, string(schemaJson))

	if err := cmd.Start(); err != nil {
		log.Printf("error starting expansion process: %s", err)
		return "", err
	}

	cmd.Wait()

	log.Printf("Expansion process: pid: %d SysTime: %v UserTime: %v", cmd.ProcessState.Pid(),
		cmd.ProcessState.SystemTime(), cmd.ProcessState.UserTime())
	if stderr.String() != "" {
		return "", fmt.Errorf("error expanding chart %s: %s", chartInv.Name, stderr.String())
	}

	return stdout.String(), nil
}
