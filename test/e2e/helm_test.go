// build +e2e

package e2e

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

const (
	timeout = 180 * time.Second
	poll    = 2 * time.Second
)

var (
	repoURL           = flag.String("repo-url", "gs://kubernetes-charts-testing", "Repository URL")
	repoName          = flag.String("repo-name", "kubernetes-charts-testing", "Repository name")
	chart             = flag.String("chart", "gs://kubernetes-charts-testing/redis-2.tgz", "Chart to deploy")
	host              = flag.String("host", "", "The URL to the helm server")
	resourcifierImage = flag.String("resourcifier-image", "", "The full image name of the Docker image for resourcifier.")
	expandybirdImage  = flag.String("expandybird-image", "", "The full image name of the Docker image for expandybird.")
	managerImage      = flag.String("manager-image", "", "The full image name of the Docker image for manager.")
)

func logKubeEnv(k *KubeContext) {
	config := k.Run("config", "view", "--flatten", "--minify").Stdout()
	k.t.Logf("Kubernetes Environment\n%s", config)
}

func TestHelm(t *testing.T) {
	kube := NewKubeContext(t)
	helm := NewHelmContext(t)

	logKubeEnv(kube)

	if !kube.Running() {
		t.Fatal("Not connected to kubernetes")
	}
	t.Log(kube.Version())
	t.Log(helm.MustRun("--version").Stdout())

	helm.Host = helmHost()
	if helm.Host == "" {
		helm.Host = fmt.Sprintf("%s%s", kube.Server(), apiProxy)
	}
	t.Logf("Using host: %v", helm.Host)

	if !helm.Running() {
		t.Log("Helm is not installed")

		install(helm)

	}

	// Add repo if it does not exsit
	if !helm.MustRun("repo", "list").Contains(*repoURL) {
		t.Logf("Adding repo %s %s", *repoName, *repoURL)
		helm.MustRun("repo", "add", *repoName, *repoURL)
	}

	// Generate a name
	deploymentName := genName()

	t.Log("Executing deploy")
	helm.MustRun("deploy",
		"--properties", "namespace=e2e",
		"--name", deploymentName,
		*chart,
	)

	//TODO get pods to lookup dynamically
	if err := wait(func() bool {
		return kube.Run("get", "pods").Match("redis.*Running")
	}); err != nil {
		t.Fatal(err)
	}
	t.Log(kube.Run("get", "pods").Stdout())

	t.Log("Executing deployment list")
	if !helm.MustRun("deployment", "list").Contains(deploymentName) {
		t.Fatal("Could not list deployment")
	}

	t.Log("Executing deployment info")
	if !helm.MustRun("deployment", "info", deploymentName).Contains("Deployed") {
		t.Fatal("Could not deploy")
	}

	t.Log("Executing deployment describe")
	helm.MustRun("deployment", "describe", deploymentName)

	t.Log("Executing deployment delete")
	if !helm.MustRun("deployment", "rm", deploymentName).Contains("Deleted") {
		t.Fatal("Could not delete deployment")
	}
}

type conditionFunc func() bool

func wait(fn conditionFunc) error {
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
		if fn() {
			return nil
		}
	}
	return fmt.Errorf("Polling timeout")
}

func genName() string {
	return fmt.Sprintf("e2e-%d", rand.Uint32())
}

func helmHost() string {
	if *host != "" {
		return *host
	}
	return os.Getenv("HELM_HOST")
}

func install(h *HelmContext) {
	args := []string{"server", "install"}
	if *expandybirdImage != "" {
		args = append(args, *expandybirdImage)
	}
	if *managerImage != "" {
		args = append(args, *managerImage)
	}
	if *resourcifierImage != "" {
		args = append(args, *resourcifierImage)
	}

	h.MustRun(args...)
	if err := wait(h.Running); err != nil {
		h.t.Fatal(err)
	}
}
