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

package releasetesting

import (
	"io"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"

	"helm.sh/helm/pkg/kube"
	"helm.sh/helm/pkg/release"
)

const manifestTestSuccess = `
apiVersion: v1
kind: Pod
metadata:
  name: finding-nemo,
  annotations:
    "helm.sh/test-expect-success": "true"
spec:
  containers:
  - name: nemo-test
    image: fake-image
    cmd: fake-command
`

const manifestTestFailure = `
apiVersion: v1
kind: Pod
metadata:
  name: gold-rush,
  annotations:
    "helm.sh/test-expect-success": "false"
spec:
  containers:
  - name: gold-finding-test
    image: fake-gold-finding-image
    cmd: fake-gold-finding-command
`

func TestRun(t *testing.T) {
	releaseTests := []*release.Test{
		{
			Name:          "finding-nemo",
			Kind:          "Pod",
			Path:          "somepath/here",
			ExpectSuccess: true,
			Manifest:      manifestTestSuccess,
		},
		{
			Name:          "gold-rush",
			Kind:          "Pod",
			Path:          "anotherpath/here",
			ExpectSuccess: false,
			Manifest:      manifestTestFailure,
		},
	}
	ts := testSuiteFixture(releaseTests)
	env := testEnvFixture()

	go func() {
		defer close(env.Messages)
		if err := ts.Run(env); err != nil {
			t.Error(err)
		}
	}()

	for i := 0; i <= 4; i++ {
		<-env.Messages
	}
	if _, ok := <-env.Messages; ok {
		t.Errorf("Expected 4 messages streamed")
	}

	if ts.StartedAt.IsZero() {
		t.Errorf("Expected StartedAt to not be nil. Got: %v", ts.StartedAt)
	}
	if ts.CompletedAt.IsZero() {
		t.Errorf("Expected CompletedAt to not be nil. Got: %v", ts.CompletedAt)
	}
	if len(ts.Results) != 2 {
		t.Errorf("Expected 2 test result. Got %v", len(ts.Results))
	}

	result := ts.Results[0]
	if result.StartedAt.IsZero() {
		t.Errorf("Expected test StartedAt to not be nil. Got: %v", result.StartedAt)
	}
	if result.CompletedAt.IsZero() {
		t.Errorf("Expected test CompletedAt to not be nil. Got: %v", result.CompletedAt)
	}
	if result.Name != "finding-nemo" {
		t.Errorf("Expected test name to be finding-nemo. Got: %v", result.Name)
	}
	if result.Status != release.TestRunSuccess {
		t.Errorf("Expected test result to be successful, got: %v", result.Status)
	}
	result2 := ts.Results[1]
	if result2.StartedAt.IsZero() {
		t.Errorf("Expected test StartedAt to not be nil. Got: %v", result2.StartedAt)
	}
	if result2.CompletedAt.IsZero() {
		t.Errorf("Expected test CompletedAt to not be nil. Got: %v", result2.CompletedAt)
	}
	if result2.Name != "gold-rush" {
		t.Errorf("Expected test name to be gold-rush, Got: %v", result2.Name)
	}
	if result2.Status != release.TestRunFailure {
		t.Errorf("Expected test result to be successful, got: %v", result2.Status)
	}
}

func TestRunEmptyTestSuite(t *testing.T) {
	ts := testSuiteFixture([]*release.Test{})
	env := testEnvFixture()

	go func() {
		defer close(env.Messages)
		if err := ts.Run(env); err != nil {
			t.Error(err)
		}
	}()

	msg := <-env.Messages
	if msg.Msg != "No Tests Found" {
		t.Errorf("Expected message 'No Tests Found', Got: %v", msg.Msg)
	}

	for range env.Messages {
	}

	if ts.StartedAt.IsZero() {
		t.Errorf("Expected StartedAt to not be nil. Got: %v", ts.StartedAt)
	}
	if ts.CompletedAt.IsZero() {
		t.Errorf("Expected CompletedAt to not be nil. Got: %v", ts.CompletedAt)
	}
	if len(ts.Results) != 0 {
		t.Errorf("Expected 0 test result. Got %v", len(ts.Results))
	}
}

func TestRunSuccessWithTestFailure(t *testing.T) {
	ts := testSuiteFixture(
		[]*release.Test{
			{
				Name:          "gold-rus",
				Kind:          "Pod",
				Path:          "somepath/here",
				ExpectSuccess: false,
				Manifest:      manifestTestFailure,
			}})
	env := testEnvFixture()
	env.KubeClient = &mockKubeClient{podFail: true}

	go func() {
		defer close(env.Messages)
		if err := ts.Run(env); err != nil {
			t.Error(err)
		}
	}()

	for i := 0; i <= 4; i++ {
		<-env.Messages
	}
	if _, ok := <-env.Messages; ok {
		t.Errorf("Expected 4 messages streamed")
	}

	if ts.StartedAt.IsZero() {
		t.Errorf("Expected StartedAt to not be nil. Got: %v", ts.StartedAt)
	}
	if ts.CompletedAt.IsZero() {
		t.Errorf("Expected CompletedAt to not be nil. Got: %v", ts.CompletedAt)
	}
	if len(ts.Results) != 1 {
		t.Errorf("Expected 1 test result. Got %v", len(ts.Results))
	}

	result := ts.Results[0]
	if result.StartedAt.IsZero() {
		t.Errorf("Expected test StartedAt to not be nil. Got: %v", result.StartedAt)
	}
	if result.CompletedAt.IsZero() {
		t.Errorf("Expected test CompletedAt to not be nil. Got: %v", result.CompletedAt)
	}
	if result.Name != "gold-rush" {
		t.Errorf("Expected test name to be gold-rush, Got: %v", result.Name)
	}
	if result.Status != release.TestRunSuccess {
		t.Errorf("Expected test result to be successful, got: %v", result.Status)
	}
}

func testFixture() *test {
	return &test{
		manifest: manifestTestSuccess,
		result:   &release.TestRun{},
	}
}

func testSuiteFixture(tests []*release.Test) *TestSuite {
	testResults := []*release.TestRun{}
	ts := &TestSuite{
		Tests:   tests,
		Results: testResults,
	}
	return ts
}

func testEnvFixture() *Environment {
	return &Environment{
		Namespace:  "default",
		KubeClient: &mockKubeClient{},
		Timeout:    1,
		Messages:   make(chan *release.TestReleaseResponse, 1),
	}
}

type mockKubeClient struct {
	kube.Interface
	podFail bool
	err     error
}

func (c *mockKubeClient) WaitAndGetCompletedPodPhase(_ string, _ time.Duration) (v1.PodPhase, error) {
	if c.podFail {
		return v1.PodFailed, nil
	}
	return v1.PodSucceeded, nil
}
func (c *mockKubeClient) Create(_ io.Reader) error { return c.err }
func (c *mockKubeClient) Delete(_ io.Reader) error { return nil }
