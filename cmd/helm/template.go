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

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/release"

	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/releaseutil"
)

const templateDesc = `
Render chart templates locally and display the output.

Any values that would normally be looked up or retrieved in-cluster will be
faked locally. Additionally, none of the server-side testing of chart validity
(e.g. whether an API is supported) is done.
`

func newTemplateCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	var validate, includeCrds, skipTests, useReleaseName bool
	var outputDir string
	client := action.NewInstall(cfg)
	valueOpts := &values.Options{}
	var kubeVersion string
	var extraAPIs []string
	var showFiles []string

	cmd := &cobra.Command{
		Use:   "template [NAME] [CHART]",
		Short: "locally render templates",
		Long:  templateDesc,
		Args:  require.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return compInstall(args, toComplete, client)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if kubeVersion != "" {
				parsedKubeVersion, err := chartutil.ParseKubeVersion(kubeVersion)
				if err != nil {
					return fmt.Errorf("invalid kube version '%s': %s", kubeVersion, err)
				}
				client.KubeVersion = parsedKubeVersion
			}

			client.DryRun = true
			client.ReleaseName = "release-name"
			client.Replace = true // Skip the name check
			client.ClientOnly = !validate
			client.APIVersions = extraAPIs
			rel, err := runInstall(args, client, valueOpts, out)
			if rel == nil || (err != nil && !settings.Debug) {
				if rel != nil {
					return fmt.Errorf("%w\n\nUse --debug flag to render out invalid YAML", err)
				}
				return err
			}
			// We ignore a potential error here because, when the --debug flag was specified,
			// we always want to print the YAML, even if it is not valid. The error is still returned afterwards.

			if outputDir != "" && useReleaseName {
				outputDir = filepath.Join(outputDir, client.ReleaseName)
			}
			fileWritten := make(map[string]bool)

			// deal crds
			if includeCrds && !client.SkipCRDs && rel.Chart != nil {
				for _, crd := range rel.Chart.CRDObjects() {
					if len(showFiles) > 0 && !matchFilePatterns(crd.Name, showFiles) {
						continue
					}
					err := writeManifest(outputDir, filepath.ToSlash(crd.Filename), string(crd.File.Data), fileWritten, true, out)
					if err != nil {
						return err
					}
				}
			}

			// deal manifests
			var manifests bytes.Buffer
			_, _ = fmt.Fprintln(&manifests, strings.TrimSpace(rel.Manifest))
			// This is necessary to ensure consistent manifest ordering when using --show-only
			// with globs or directory names.
			splitManifests := releaseutil.SplitManifests(manifests.String())
			manifestsKeys := make([]string, 0, len(splitManifests))
			// such as `# Source: subchart/templates/service.yaml` will be divided into two parts `subchart/`, `templates/service.yaml`
			// and manifestName will be `templates/service.yaml` , manifestPath will be `subchart/templates/service.yam`
			manifestNameRegex := regexp.MustCompile("# Source: ([^/]+/)(.+)")
			for k := range splitManifests {
				manifestsKeys = append(manifestsKeys, k)
			}
			sort.Sort(releaseutil.BySplitManifestsOrder(manifestsKeys))
			for _, manifestKey := range manifestsKeys {
				manifest := splitManifests[manifestKey]
				submatch := manifestNameRegex.FindStringSubmatch(manifest)
				var manifestName, manifestPath string
				if len(submatch) > 2 {
					manifestName = submatch[2]
					manifestPath = submatch[1] + submatch[2]
				}
				if len(showFiles) > 0 && !matchFilePatterns(manifestName, showFiles) {
					continue
				}
				err := writeManifest(outputDir, manifestPath, manifest, fileWritten, false, out)
				if err != nil {
					return err
				}
			}

			// deal hooks
			if !client.DisableHooks {
				for _, m := range rel.Hooks {
					if (skipTests && isTestHook(m)) || (len(showFiles) > 0 && !matchFilePatterns(m.Name, showFiles)) {
						continue
					}
					err := writeManifest(outputDir, m.Path, m.Manifest, fileWritten, true, out)
					if err != nil {
						return err
					}
				}
			}
			return err
		},
	}

	f := cmd.Flags()
	addInstallFlags(cmd, f, client, valueOpts)
	f.StringArrayVarP(&showFiles, "show-only", "s", []string{}, "only show manifests rendered from the given templates")
	f.StringVar(&outputDir, "output-dir", "", "writes the executed templates to files in output-dir instead of stdout")
	f.BoolVar(&validate, "validate", false, "validate your manifests against the Kubernetes cluster you are currently pointing at. This is the same validation performed on an install")
	f.BoolVar(&includeCrds, "include-crds", false, "include CRDs in the templated output")
	f.BoolVar(&skipTests, "skip-tests", false, "skip tests from templated output")
	f.BoolVar(&client.IsUpgrade, "is-upgrade", false, "set .Release.IsUpgrade instead of .Release.IsInstall")
	f.StringVar(&kubeVersion, "kube-version", "", "Kubernetes version used for Capabilities.KubeVersion")
	f.StringSliceVarP(&extraAPIs, "api-versions", "a", []string{}, "Kubernetes api versions used for Capabilities.APIVersions")
	f.BoolVar(&useReleaseName, "release-name", false, "use release name in the output-dir path.")
	bindPostRenderFlag(cmd, &client.PostRenderer)

	return cmd
}

func isTestHook(h *release.Hook) bool {
	for _, e := range h.Events {
		if e == release.HookTest {
			return true
		}
	}
	return false
}

// writeToFile write manifests into output dir.
func writeToFile(outputDir string, name string, data string, append, withHeader bool) error {
	outfileName := filepath.Join(outputDir, name)

	err := ensureDirectoryForFile(outfileName)
	if err != nil {
		return err
	}

	f, err := createOrOpenFile(outfileName, append)
	if err != nil {
		return err
	}

	defer f.Close()

	err = writeStream(name, data, withHeader, f)

	if err != nil {
		return err
	}

	fmt.Printf("wrote %s\n", outfileName)
	return nil
}

func createOrOpenFile(filename string, append bool) (*os.File, error) {
	if append {
		return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	}
	return os.Create(filename)
}

func ensureDirectoryForFile(file string) error {
	baseDir := path.Dir(file)
	_, err := os.Stat(baseDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(baseDir, 0755)
}

func matchFilePatterns(target string, sf []string) bool {
	for _, pattern := range sf {
		pattern = filepath.ToSlash(pattern)
		matched, _ := filepath.Match(pattern, target)
		if matched {
			return true
		}
	}
	return false
}

// writeManifest write manifest to stdout or file stream. use withHeader to control if write file header `# Source: XXXXX.yaml`.
func writeManifest(outputDir, path, manifest string, fileWritten map[string]bool, withHeader bool, outStream io.Writer) error {
	if outputDir == "" {
		return writeStream(path, manifest, withHeader, outStream)
	} else {
		err := writeToFile(outputDir, path, manifest, fileWritten[path], withHeader)
		if err != nil {
			return err
		}
		fileWritten[path] = true
	}
	return nil
}

func writeStream(path, manifest string, withHeader bool, outStream io.Writer) error {
	//write yaml delimiter
	_, err := fmt.Fprintf(outStream, "---\n")
	if err != nil {
		return err
	}
	//write file header
	if withHeader {
		_, err = fmt.Fprintf(outStream, "# Source: %s\n", path)
		if err != nil {
			return err
		}
	}

	//write manifest content
	_, err = fmt.Fprintf(outStream, "%s\n", manifest)
	return err
}
