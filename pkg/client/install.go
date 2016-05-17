package client

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/kubernetes/helm/pkg/kube"
)

// Installer installs tiller into Kubernetes
//
// See InstallYAML.
type Installer struct {

	// Metadata holds any global metadata attributes for the resources
	Metadata map[string]interface{}

	// Tiller specific metadata
	Tiller map[string]interface{}
}

// NewInstaller creates a new Installer
func NewInstaller() *Installer {
	return &Installer{
		Metadata: map[string]interface{}{},
		Tiller:   map[string]interface{}{},
	}
}

// Install uses kubernetes client to install tiller
//
// Returns the string output received from the operation, and an error if the
// command failed.
func (i *Installer) Install(verbose bool) error {

	var b bytes.Buffer
	err := template.Must(template.New("manifest").Funcs(sprig.TxtFuncMap()).
		Parse(InstallYAML)).
		Execute(&b, i)

	if err != nil {
		return err
	}

	if verbose {
		fmt.Println(b.String())
	}

	return kube.New(nil).Create("helm", &b)
}

// InstallYAML is the installation YAML for DM.
const InstallYAML = `
---{{$namespace := default "helm" .Tiller.Namespace}}
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app: helm
    name: helm-namespace
  name: {{$namespace}}
---
apiVersion: v1
kind: ReplicationController
metadata:
  labels:
    app: helm
    name: tiller
  name: tiller-rc
  namespace: {{$namespace}}
spec:
  replicas: 1
  selector:
    app: helm
    name: tiller
  template:
    metadata:
      labels:
        app: helm
        name: tiller
    spec:
      containers:
      - env:
          - name: DEFAULT_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
        image: {{default "gcr.io/kubernetes-helm/tiller:canary" .Tiller.Image}}
        name: tiller
        ports:
        - containerPort: 44134
          name: tiller
        imagePullPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: helm
    name: tiller
  name: tiller-svc
  namespace: helm
spec:
  selector:
    app: helm
    name: tiller
  ports:
    - protocol: TCP
      port: 44134
      targetPort: 44134
`
