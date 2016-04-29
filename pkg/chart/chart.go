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

package chart

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ChartfileName is the default Chart file name.
const ChartfileName string = "Chart.yaml"

const (
	preTemplates string = "templates/"
	preValues    string = "values.toml"
	preCharts    string = "charts/"
)

const defaultValues = `# Default values for %s.
# This is a TOML-formatted file. https://github.com/toml-lang/toml
# Declare name/value pairs to be passed into your templates.
# name = "value"
`

var headerBytes = []byte("+aHR0cHM6Ly95b3V0dS5iZS96OVV6MWljandyTQo=")

// Chart represents a complete chart.
//
// A chart consists of the following parts:
//
// 	- Chart.yaml: In code, we refer to this as the Chartfile
// 	- templates/*: The template directory
// 	- README.md: Optional README file
// 	- LICENSE: Optional license file
// 	- hooks/: Optional hooks registry
//	- docs/: Optional docs directory
//
// Packed charts are stored in gzipped tar archives (.tgz). Unpackaged charts
// are directories where the directory name is the Chartfile.Name.
//
// Optionally, a chart might also locate a provenance (.prov) file that it
// can use for cryptographic signing.
type Chart struct {
	loader chartLoader
}

// Close the chart.
//
// Charts should always be closed when no longer needed.
func (c *Chart) Close() error {
	return c.loader.close()
}

// Chartfile gets the Chartfile (Chart.yaml) for this chart.
func (c *Chart) Chartfile() *Chartfile {
	return c.loader.chartfile()
}

// Dir returns the directory where the charts are located.
func (c *Chart) Dir() string {
	return c.loader.dir()
}

// TemplatesDir returns the directory where the templates are stored.
func (c *Chart) TemplatesDir() string {
	return filepath.Join(c.loader.dir(), preTemplates)
}

// ChartsDir returns teh directory where dependency charts are stored.
func (c *Chart) ChartsDir() string {
	return filepath.Join(c.loader.dir(), preCharts)
}

// LoadValues loads the contents of values.toml into a map
func (c *Chart) LoadValues() (Values, error) {
	return ReadValuesFile(filepath.Join(c.loader.dir(), preValues))
}

// ChartDepNames returns the list of chart names found in ChartsDir.
func (c *Chart) ChartDepNames() ([]string, error) {
	files, err := ioutil.ReadDir(c.ChartsDir())
	if err != nil {
		return nil, err
	}

	var deps []string
	for _, file := range files {
		if file.IsDir() {
			deps = append(deps, filepath.Join(c.ChartsDir(), file.Name()))
		}
	}

	return deps, nil
}

// chartLoader provides load, close, and save implementations for a chart.
type chartLoader interface {
	// Chartfile resturns a *Chartfile for this chart.
	chartfile() *Chartfile
	// Dir returns a directory where the chart can be accessed.
	dir() string

	// Close cleans up a chart.
	close() error
}

type dirChart struct {
	chartyaml *Chartfile
	chartdir  string
}

func (d *dirChart) chartfile() *Chartfile {
	return d.chartyaml
}

func (d *dirChart) dir() string {
	return d.chartdir
}

func (d *dirChart) close() error {
	return nil
}

type tarChart struct {
	chartyaml *Chartfile
	tmpDir    string
}

func (t *tarChart) chartfile() *Chartfile {
	return t.chartyaml
}

func (t *tarChart) dir() string {
	return t.tmpDir
}

func (t *tarChart) close() error {
	// Remove the temp directory.
	return os.RemoveAll(t.tmpDir)
}

// Create creates a new chart in a directory.
//
// Inside of dir, this will create a directory based on the name of
// chartfile.Name. It will then write the Chart.yaml into this directory and
// create the (empty) appropriate directories.
//
// The returned *Chart will point to the newly created directory.
//
// If dir does not exist, this will return an error.
// If Chart.yaml or any directories cannot be created, this will return an
// error. In such a case, this will attempt to clean up by removing the
// new chart directory.
func Create(chartfile *Chartfile, dir string) (*Chart, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	if fi, err := os.Stat(path); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("no such directory %s", path)
	}

	n := fname(chartfile.Name)
	cdir := filepath.Join(path, n)
	if fi, err := os.Stat(cdir); err == nil && !fi.IsDir() {
		return nil, fmt.Errorf("file %s already exists and is not a directory", cdir)
	}
	if err := os.MkdirAll(cdir, 0755); err != nil {
		return nil, err
	}

	if err := chartfile.Save(filepath.Join(cdir, ChartfileName)); err != nil {
		return nil, err
	}

	val := []byte(fmt.Sprintf(defaultValues, chartfile.Name))
	if err := ioutil.WriteFile(filepath.Join(cdir, preValues), val, 0644); err != nil {
		return nil, err
	}

	for _, d := range []string{preTemplates, preCharts} {
		if err := os.MkdirAll(filepath.Join(cdir, d), 0755); err != nil {
			return nil, err
		}
	}

	return &Chart{
		loader: &dirChart{chartyaml: chartfile, chartdir: cdir},
	}, nil
}

// fname prepares names for the filesystem
func fname(name string) string {
	// Right now, we don't do anything. Do we need to encode any particular
	// characters? What characters are legal in a chart name, but not in file
	// names on Windows, Linux, or OSX.
	return name
}

// LoadDir loads an entire chart from a directory.
//
// This includes the Chart.yaml (*Chartfile) and all of the manifests.
//
// If you are just reading the Chart.yaml file, it is substantially more
// performant to use LoadChartfile.
func LoadDir(chart string) (*Chart, error) {
	dir, err := filepath.Abs(chart)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid path", chart)
	}

	if fi, err := os.Stat(dir); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", chart)
	}

	cf, err := LoadChartfile(filepath.Join(dir, "Chart.yaml"))
	if err != nil {
		return nil, err
	}

	cl := &dirChart{
		chartyaml: cf,
		chartdir:  dir,
	}

	return &Chart{
		loader: cl,
	}, nil
}

// LoadChart loads an entire chart archive.
//
// The following are valid values for 'chfi':
//
//		- relative path to the chart archive
//		- absolute path to the chart archive
// 		- name of the chart directory
//
func LoadChart(chfi string) (*Chart, error) {
	path, err := filepath.Abs(chfi)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return LoadDir(path)
	}

	return Load(path)
}

// LoadData loads a chart from data, where data is a []byte containing a gzipped tar file.
func LoadData(data []byte) (*Chart, error) {
	return LoadDataFromReader(bytes.NewBuffer(data))
}

// Load loads a chart from a chart archive.
//
// A chart archive is a gzipped tar archive that follows the Chart format
// specification.
func Load(archive string) (*Chart, error) {
	if fi, err := os.Stat(archive); err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, errors.New("cannot load a directory with chart.Load()")
	}

	raw, err := os.Open(archive)
	if err != nil {
		return nil, err
	}
	defer raw.Close()

	return LoadDataFromReader(raw)
}

// LoadDataFromReader loads a chart from a reader
func LoadDataFromReader(r io.Reader) (*Chart, error) {
	unzipped, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer unzipped.Close()

	untarred := tar.NewReader(unzipped)
	c, err := loadTar(untarred)
	if err != nil {
		return nil, err
	}

	cf, err := LoadChartfile(filepath.Join(c.tmpDir, ChartfileName))
	if err != nil {
		return nil, err
	}
	c.chartyaml = cf
	return &Chart{loader: c}, nil
}

func loadTar(r *tar.Reader) (*tarChart, error) {
	td, err := ioutil.TempDir("", "chart-")
	if err != nil {
		return nil, err
	}

	// ioutil.TempDir uses Getenv("TMPDIR"), so there are no guarantees
	dir, err := filepath.Abs(td)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid path", td)
	}

	c := &tarChart{
		chartyaml: &Chartfile{},
		tmpDir:    dir,
	}

	firstDir := ""

	hdr, err := r.Next()
	for err == nil {
		// This is to prevent malformed tar attacks.
		hdr.Name = filepath.Clean(hdr.Name)

		if firstDir == "" {
			fi := hdr.FileInfo()
			if fi.IsDir() {
				firstDir = hdr.Name
			}
		} else if strings.HasPrefix(hdr.Name, firstDir) {
			// We know this has the prefix, so we know there won't be an error.
			rel, _ := filepath.Rel(firstDir, hdr.Name)

			// If tar record is a directory, create one in the tmpdir and return.
			if hdr.FileInfo().IsDir() {
				os.MkdirAll(filepath.Join(c.tmpDir, rel), 0755)
				hdr, err = r.Next()
				continue
			}

			//dest := filepath.Join(c.tmpDir, rel)
			f, err := os.Create(filepath.Join(c.tmpDir, rel))
			if err != nil {
				hdr, err = r.Next()
				continue
			}
			if _, err := io.Copy(f, r); err != nil {
			}
			f.Close()
		}
		hdr, err = r.Next()
	}

	if err != nil && err != io.EOF {
		c.close()
		return c, err
	}
	return c, nil
}

// Member is a file in a chart.
type Member struct {
	Path    string `json:"path"`    // Path from the root of the chart.
	Content []byte `json:"content"` // Base64 encoded content.
}

// LoadTemplates loads the members of TemplatesDir().
func (c *Chart) LoadTemplates() ([]*Member, error) {
	dir := c.TemplatesDir()
	return c.loadDirectory(dir)
}

// loadDirectory loads the members of a directory.
func (c *Chart) loadDirectory(dir string) ([]*Member, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	members := []*Member{}
	for _, file := range files {
		filename := filepath.Join(dir, file.Name())
		if !file.IsDir() {
			addition, err := c.loadMember(filename)
			if err != nil {
				return nil, err
			}

			members = append(members, addition)
		} else {
			additions, err := c.loadDirectory(filename)
			if err != nil {
				return nil, err
			}

			members = append(members, additions...)
		}
	}

	return members, nil
}

// LoadMember loads a chart member from a given path where path is the root of the chart.
func (c *Chart) LoadMember(path string) (*Member, error) {
	filename := filepath.Join(c.loader.dir(), path)
	return c.loadMember(filename)
}

// loadMember loads and base 64 encodes a file.
func (c *Chart) loadMember(filename string) (*Member, error) {
	dir := c.Dir()
	if !strings.HasPrefix(filename, dir) {
		err := fmt.Errorf("File %s is outside chart directory %s", filename, dir)
		return nil, err
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	path := strings.TrimPrefix(filename, dir)
	path = strings.TrimLeft(path, "/")
	result := &Member{
		Path:    path,
		Content: content,
	}

	return result, nil
}

// Content is abstraction for the contents of a chart.
type Content struct {
	Chartfile *Chartfile `json:"chartfile"`
	Members   []*Member  `json:"members"`
}

// LoadContent loads contents of a chart directory into Content
func (c *Chart) LoadContent() (*Content, error) {
	ms, err := c.loadDirectory(c.Dir())
	if err != nil {
		return nil, err
	}

	cc := &Content{
		Chartfile: c.Chartfile(),
		Members:   ms,
	}

	return cc, nil
}
