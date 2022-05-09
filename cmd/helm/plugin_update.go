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
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/pkg/plugin"
	"helm.sh/helm/v3/pkg/plugin/installer"
)

type pluginUpdateOptions struct {
	names   map[string]string
	version string
}

const pluginUpdateDesc = `
This command allows you to update one or more Helm plugins.
`

func newPluginUpdateCmd(out io.Writer) *cobra.Command {
	o := &pluginUpdateOptions{}

	cmd := &cobra.Command{
		Use:     "update <plugin:[version]>...",
		Aliases: []string{"up"},
		Short:   "update one or more Helm plugins",
		Long:    pluginUpdateDesc,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return compListPlugins(toComplete, args), cobra.ShellCompDirectiveNoFileComp
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return o.complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}
	cmd.Flags().StringVar(&o.version, "version", "", "specify a version constraint. If this is not specified, the latest version is installed")
	return cmd
}

func (o *pluginUpdateOptions) complete(args []string) error {
	if len(args) == 0 {
		return errors.New("please provide plugin name to update")
	}
	o.names = make(map[string]string)
	for _, arg := range args {
		if strings.ContainsAny(arg, ":") {
			o.names[arg[:strings.LastIndex(arg, ":")]] = arg[strings.LastIndex(arg, ":")+1:]
		} else {
			o.names[arg] = ""
		}
	}
	return nil
}

func (o *pluginUpdateOptions) run(out io.Writer) error {
	installer.Debug = settings.Debug
	debug("loading installed plugins from %s", settings.PluginsDirectory)
	plugins, err := plugin.FindPlugins(settings.PluginsDirectory)
	if err != nil {
		return err
	}
	var errorPlugins []string
	for name, version := range o.names {
		if found := findPlugin(plugins, name); found != nil {
			if err := updatePlugin(found, version); err != nil {
				errorPlugins = append(errorPlugins, fmt.Sprintf("Failed to update plugin %s, got error (%v)", name, err))
			} else {
				fmt.Fprintf(out, "Updated plugin: %s\n", name)
			}
		} else {
			errorPlugins = append(errorPlugins, fmt.Sprintf("Plugin: %s not found", name))
		}
	}
	if len(errorPlugins) > 0 {
		return errors.Errorf(strings.Join(errorPlugins, "\n"))
	}
	return nil
}

func updatePlugin(p *plugin.Plugin, version string) error {
	exactLocation, err := filepath.EvalSymlinks(p.Dir)
	if err != nil {
		return err
	}
	absExactLocation, err := filepath.Abs(exactLocation)
	if err != nil {
		return err
	}

	i, err := installer.FindSource(absExactLocation, version)
	if err != nil {
		return err
	}
	if err := installer.Update(i); err != nil {
		return err
	}

	debug("loading plugin from %s", i.Path())
	updatedPlugin, err := plugin.LoadDir(i.Path())
	if err != nil {
		return err
	}

	return runHook(updatedPlugin, plugin.Update)
}
