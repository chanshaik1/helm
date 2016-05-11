package kube

import (
	"fmt"
	"io"

	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

// Client represents a client capable of communicating with the Kubernetes API.
type Client struct {
	config clientcmd.ClientConfig
}

// New create a new Client
func New(config clientcmd.ClientConfig) *Client {
	return &Client{
		config: config,
	}
}

// ResourceActorFunc performs an action on a signle resource.
type ResourceActorFunc func(*resource.Info) error

// Create creates kubernetes resources from an io.reader
//
// Namespace will set the namespace
func (c *Client) Create(namespace string, reader io.Reader) error {
	return perform(c, namespace, reader, createResource)
}

// Delete deletes kubernetes resources from an io.reader
//
// Namespace will set the namespace
func (c *Client) Delete(namespace string, reader io.Reader) error {
	return perform(c, namespace, reader, deleteResource)
}

func (c *Client) factory() *cmdutil.Factory {
	return cmdutil.NewFactory(c.config)
}

const includeThirdPartyAPIs = false

func perform(c *Client, namespace string, reader io.Reader, fn ResourceActorFunc) error {
	r := c.factory().NewBuilder(includeThirdPartyAPIs).
		ContinueOnError().
		NamespaceParam(namespace).
		RequireNamespace().
		Stream(reader, "").
		Flatten().
		Do()

	if r.Err() != nil {
		return r.Err()
	}

	count := 0
	err := r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}
		err = fn(info)

		if err == nil {
			count++
		}
		return err
	})

	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no objects passed to create")
	}
	return nil
}

func createResource(info *resource.Info) error {
	_, err := resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object)
	return err
}

func deleteResource(info *resource.Info) error {
	return resource.NewHelper(info.Client, info.Mapping).Delete(info.Namespace, info.Name)
}
