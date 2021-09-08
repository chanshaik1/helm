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
	"strings"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/repo"
)

type repoListOptions struct {
	allowEmpty   bool
	repoFile     string
	outputFormat output.Format
}

func newRepoListCmd(out io.Writer) *cobra.Command {
	o := &repoListOptions{}

	cmd := &cobra.Command{
		Use:               "list",
		Aliases:           []string{"ls"},
		Short:             "list chart repositories",
		Args:              require.NoArgs,
		ValidArgsFunction: noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.repoFile = settings.RepositoryConfig
			return o.run(out)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&o.allowEmpty, "allow-empty", false, "exit with status 0 and no error message if no repositories to show")
	bindOutputFlag(cmd, &o.outputFormat)

	return cmd
}

func (o *repoListOptions) run(out io.Writer) error {
	f, err := repo.LoadFile(o.repoFile)
	if o.allowEmpty {
		return o.outputFormat.Write(out, &repoListWriter{f.Repositories})
	}
	if isNotExist(err) || (len(f.Repositories) == 0 && !(o.outputFormat == output.JSON || o.outputFormat == output.YAML)) {
		return errors.New("no repositories to show")
	}
	return o.outputFormat.Write(out, &repoListWriter{f.Repositories})
}

type repositoryElement struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type repoListWriter struct {
	repos []*repo.Entry
}

func (r *repoListWriter) WriteTable(out io.Writer) error {
	table := uitable.New()
	table.AddRow("NAME", "URL")
	for _, re := range r.repos {
		table.AddRow(re.Name, re.URL)
	}
	return output.EncodeTable(out, table)
}

func (r *repoListWriter) WriteJSON(out io.Writer) error {
	return r.encodeByFormat(out, output.JSON)
}

func (r *repoListWriter) WriteYAML(out io.Writer) error {
	return r.encodeByFormat(out, output.YAML)
}

func (r *repoListWriter) encodeByFormat(out io.Writer, format output.Format) error {
	// Initialize the array so no results returns an empty array instead of null
	repolist := make([]repositoryElement, 0, len(r.repos))

	for _, re := range r.repos {
		repolist = append(repolist, repositoryElement{Name: re.Name, URL: re.URL})
	}

	switch format {
	case output.JSON:
		return output.EncodeJSON(out, repolist)
	case output.YAML:
		return output.EncodeYAML(out, repolist)
	}

	// Because this is a non-exported function and only called internally by
	// WriteJSON and WriteYAML, we shouldn't get invalid types
	return nil
}

// Returns all repos from repos, except those with names matching ignoredRepoNames
// Inspired by https://stackoverflow.com/a/28701031/893211
func filterRepos(repos []*repo.Entry, ignoredRepoNames []string) []*repo.Entry {
	// if ignoredRepoNames is nil, just return repo
	if ignoredRepoNames == nil {
		return repos
	}

	filteredRepos := make([]*repo.Entry, 0)

	ignored := make(map[string]bool, len(ignoredRepoNames))
	for _, repoName := range ignoredRepoNames {
		ignored[repoName] = true
	}

	for _, repo := range repos {
		if _, removed := ignored[repo.Name]; !removed {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos
}

// Provide dynamic auto-completion for repo names
func compListRepos(prefix string, ignoredRepoNames []string) []string {
	var rNames []string

	f, err := repo.LoadFile(settings.RepositoryConfig)
	if err == nil && len(f.Repositories) > 0 {
		filteredRepos := filterRepos(f.Repositories, ignoredRepoNames)
		for _, repo := range filteredRepos {
			if strings.HasPrefix(repo.Name, prefix) {
				rNames = append(rNames, fmt.Sprintf("%s\t%s", repo.Name, repo.URL))
			}
		}
	}
	return rNames
}
