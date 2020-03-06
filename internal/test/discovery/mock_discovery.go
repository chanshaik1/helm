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

// Code generated by MockGen. DO NOT EDIT.
// Source: k8s.io/client-go/discovery (interfaces: CachedDiscoveryInterface)
// Command used to generate: mockgen -package discovery -destination internal/test/discovery/mock_discovery.go k8s.io/client-go/discovery CachedDiscoveryInterface

// Package discovery is a generated GoMock package.
package discovery

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	version "k8s.io/apimachinery/pkg/version"
	rest "k8s.io/client-go/rest"
)

// MockCachedDiscoveryInterface is a mock of CachedDiscoveryInterface interface
type MockCachedDiscoveryInterface struct {
	ctrl     *gomock.Controller
	recorder *MockCachedDiscoveryInterfaceMockRecorder
}

// MockCachedDiscoveryInterfaceMockRecorder is the mock recorder for MockCachedDiscoveryInterface
type MockCachedDiscoveryInterfaceMockRecorder struct {
	mock *MockCachedDiscoveryInterface
}

// NewMockCachedDiscoveryInterface creates a new mock instance
func NewMockCachedDiscoveryInterface(ctrl *gomock.Controller) *MockCachedDiscoveryInterface {
	mock := &MockCachedDiscoveryInterface{ctrl: ctrl}
	mock.recorder = &MockCachedDiscoveryInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCachedDiscoveryInterface) EXPECT() *MockCachedDiscoveryInterfaceMockRecorder {
	return m.recorder
}

// Fresh mocks base method
func (m *MockCachedDiscoveryInterface) Fresh() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fresh")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Fresh indicates an expected call of Fresh
func (mr *MockCachedDiscoveryInterfaceMockRecorder) Fresh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fresh", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).Fresh))
}

// Invalidate mocks base method
func (m *MockCachedDiscoveryInterface) Invalidate() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Invalidate")
}

// Invalidate indicates an expected call of Invalidate
func (mr *MockCachedDiscoveryInterfaceMockRecorder) Invalidate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Invalidate", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).Invalidate))
}

// OpenAPISchema mocks base method
func (m *MockCachedDiscoveryInterface) OpenAPISchema() (*openapi_v2.Document, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenAPISchema")
	ret0, _ := ret[0].(*openapi_v2.Document)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenAPISchema indicates an expected call of OpenAPISchema
func (mr *MockCachedDiscoveryInterfaceMockRecorder) OpenAPISchema() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenAPISchema", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).OpenAPISchema))
}

// RESTClient mocks base method
func (m *MockCachedDiscoveryInterface) RESTClient() rest.Interface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RESTClient")
	ret0, _ := ret[0].(rest.Interface)
	return ret0
}

// RESTClient indicates an expected call of RESTClient
func (mr *MockCachedDiscoveryInterfaceMockRecorder) RESTClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RESTClient", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).RESTClient))
}

// ServerGroups mocks base method
func (m *MockCachedDiscoveryInterface) ServerGroups() (*v1.APIGroupList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerGroups")
	ret0, _ := ret[0].(*v1.APIGroupList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServerGroups indicates an expected call of ServerGroups
func (mr *MockCachedDiscoveryInterfaceMockRecorder) ServerGroups() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerGroups", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).ServerGroups))
}

// ServerGroupsAndResources mocks base method
func (m *MockCachedDiscoveryInterface) ServerGroupsAndResources() ([]*v1.APIGroup, []*v1.APIResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerGroupsAndResources")
	ret0, _ := ret[0].([]*v1.APIGroup)
	ret1, _ := ret[1].([]*v1.APIResourceList)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ServerGroupsAndResources indicates an expected call of ServerGroupsAndResources
func (mr *MockCachedDiscoveryInterfaceMockRecorder) ServerGroupsAndResources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerGroupsAndResources", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).ServerGroupsAndResources))
}

// ServerPreferredNamespacedResources mocks base method
func (m *MockCachedDiscoveryInterface) ServerPreferredNamespacedResources() ([]*v1.APIResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerPreferredNamespacedResources")
	ret0, _ := ret[0].([]*v1.APIResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServerPreferredNamespacedResources indicates an expected call of ServerPreferredNamespacedResources
func (mr *MockCachedDiscoveryInterfaceMockRecorder) ServerPreferredNamespacedResources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerPreferredNamespacedResources", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).ServerPreferredNamespacedResources))
}

// ServerPreferredResources mocks base method
func (m *MockCachedDiscoveryInterface) ServerPreferredResources() ([]*v1.APIResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerPreferredResources")
	ret0, _ := ret[0].([]*v1.APIResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServerPreferredResources indicates an expected call of ServerPreferredResources
func (mr *MockCachedDiscoveryInterfaceMockRecorder) ServerPreferredResources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerPreferredResources", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).ServerPreferredResources))
}

// ServerResources mocks base method
func (m *MockCachedDiscoveryInterface) ServerResources() ([]*v1.APIResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerResources")
	ret0, _ := ret[0].([]*v1.APIResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServerResources indicates an expected call of ServerResources
func (mr *MockCachedDiscoveryInterfaceMockRecorder) ServerResources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerResources", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).ServerResources))
}

// ServerResourcesForGroupVersion mocks base method
func (m *MockCachedDiscoveryInterface) ServerResourcesForGroupVersion(arg0 string) (*v1.APIResourceList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerResourcesForGroupVersion", arg0)
	ret0, _ := ret[0].(*v1.APIResourceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServerResourcesForGroupVersion indicates an expected call of ServerResourcesForGroupVersion
func (mr *MockCachedDiscoveryInterfaceMockRecorder) ServerResourcesForGroupVersion(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerResourcesForGroupVersion", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).ServerResourcesForGroupVersion), arg0)
}

// ServerVersion mocks base method
func (m *MockCachedDiscoveryInterface) ServerVersion() (*version.Info, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerVersion")
	ret0, _ := ret[0].(*version.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServerVersion indicates an expected call of ServerVersion
func (mr *MockCachedDiscoveryInterfaceMockRecorder) ServerVersion() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerVersion", reflect.TypeOf((*MockCachedDiscoveryInterface)(nil).ServerVersion))
}
