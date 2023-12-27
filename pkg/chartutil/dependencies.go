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

package chartutil

import (
	"go/build/constraint"
	"log"
	"strings"

	"github.com/mitchellh/copystructure"

	"helm.sh/helm/v3/pkg/chart"
)

// ProcessDependencies checks through this chart's dependencies, processing accordingly.
//
// TODO: For Helm v4 this can be combined with or turned into ProcessDependenciesWithMerge
func ProcessDependencies(c *chart.Chart, v Values) error {
	if err := processDependencyEnabled(c, v, ""); err != nil {
		return err
	}
	return processDependencyImportValues(c, false)
}

// ProcessDependenciesWithMerge checks through this chart's dependencies, processing accordingly.
// It is similar to ProcessDependencies but it does not remove nil values during
// the import/export handling process.
func ProcessDependenciesWithMerge(c *chart.Chart, v Values) error {
	if err := processDependencyEnabled(c, v, ""); err != nil {
		return err
	}
	return processDependencyImportValues(c, true)
}

// processDependencyConditions disables charts based on condition path value in values
//
// Adapted from git@github.com:golang/go.git src/go/build/constraint/expr.go:394
//
// In order to be compatible with the previous design, this is a bit different
// from the tag resolve logic of go.
//
// ' ' for OR, ',' for AND, '!' prefix for neg
//
// If an expr is not specified, it is not false, but null, that is, it will be ignored
//
// !cond1,cond2 cond3
// cond1 = true, cond2 = true, cond3 = true
// Usually means (false && true || true) -> true
//
// but if
// cond1 = false, cond2 = null, cond3 = null
// the result is (true && null || null) -> true
// but not (true && false || false) -> false
func processDependencyConditions(reqs []*chart.Dependency, cvals Values, cpath string) {
	if reqs == nil {
		return
	}
	for _, r := range reqs {

		var x constraint.Expr
		for _, clause := range strings.Fields(r.Condition) {
			var y constraint.Expr
			for _, lit := range strings.Split(clause, ",") {
				var z constraint.Expr
				var neg bool
				if strings.HasPrefix(lit, "!") {
					lit = lit[len("!"):]
					neg = true
				}

				// retrieve value
				vv, err := cvals.PathValue(cpath + lit)
				if err == nil {
					// if not bool, warn
					if _, ok := vv.(bool); ok {
						z = tag(lit)
						if neg {
							z = not(z)
						}
					} else {
						log.Printf("Warning: Condition path '%s' for chart %s returned non-bool value", lit, r.Name)
					}
				} else if _, ok := err.(ErrNoValue); !ok {
					// this is a real error
					log.Printf("Warning: PathValue returned error %v", err)
				}

				if z != nil {
					if y == nil {
						y = z
					} else {
						y = and(y, z)
					}
				}

			}

			if y != nil {
				if x == nil {
					x = y
				} else {
					x = or(x, y)
				}
			}

		}

		if x != nil {
			r.Enabled = x.Eval(
				func(tag string) bool {

					// This value has been checked before, so it is unlikely that an error will be reported here
					vv, _ := cvals.PathValue(cpath + tag)
					if bv, ok := vv.(bool); ok {
						return bv
					}

					return false
				},
			)
		}
	}
}

func tag(tag string) constraint.Expr           { return &constraint.TagExpr{Tag: tag} }
func not(x constraint.Expr) constraint.Expr    { return &constraint.NotExpr{X: x} }
func and(x, y constraint.Expr) constraint.Expr { return &constraint.AndExpr{X: x, Y: y} }
func or(x, y constraint.Expr) constraint.Expr  { return &constraint.OrExpr{X: x, Y: y} }

// processDependencyTags disables charts based on tags in values
func processDependencyTags(reqs []*chart.Dependency, cvals Values) {
	if reqs == nil {
		return
	}
	vt, err := cvals.Table("tags")
	if err != nil {
		return
	}
	for _, r := range reqs {
		var hasTrue, hasFalse bool
		for _, k := range r.Tags {
			if b, ok := vt[k]; ok {
				// if not bool, warn
				if bv, ok := b.(bool); ok {
					if bv {
						hasTrue = true
					} else {
						hasFalse = true
					}
				} else {
					log.Printf("Warning: Tag '%s' for chart %s returned non-bool value", k, r.Name)
				}
			}
		}
		if !hasTrue && hasFalse {
			r.Enabled = false
		} else if hasTrue || !hasTrue && !hasFalse {
			r.Enabled = true
		}
	}
}

func getAliasDependency(charts []*chart.Chart, dep *chart.Dependency) *chart.Chart {
	for _, c := range charts {
		if c == nil {
			continue
		}
		if c.Name() != dep.Name {
			continue
		}
		if !IsCompatibleRange(dep.Version, c.Metadata.Version) {
			continue
		}

		out := *c
		md := *c.Metadata
		out.Metadata = &md

		if dep.Alias != "" {
			md.Name = dep.Alias
		}
		return &out
	}
	return nil
}

// processDependencyEnabled removes disabled charts from dependencies
func processDependencyEnabled(c *chart.Chart, v map[string]interface{}, path string) error {
	if c.Metadata.Dependencies == nil {
		return nil
	}

	var chartDependencies []*chart.Chart
	// If any dependency is not a part of Chart.yaml
	// then this should be added to chartDependencies.
	// However, if the dependency is already specified in Chart.yaml
	// we should not add it, as it would be anyways processed from Chart.yaml

Loop:
	for _, existing := range c.Dependencies() {
		for _, req := range c.Metadata.Dependencies {
			if existing.Name() == req.Name && IsCompatibleRange(req.Version, existing.Metadata.Version) {
				continue Loop
			}
		}
		chartDependencies = append(chartDependencies, existing)
	}

	for _, req := range c.Metadata.Dependencies {
		if req == nil {
			continue
		}
		if chartDependency := getAliasDependency(c.Dependencies(), req); chartDependency != nil {
			chartDependencies = append(chartDependencies, chartDependency)
		}
		if req.Alias != "" {
			req.Name = req.Alias
		}
	}
	c.SetDependencies(chartDependencies...)

	// set all to true
	for _, lr := range c.Metadata.Dependencies {
		lr.Enabled = true
	}
	cvals, err := CoalesceValues(c, v)
	if err != nil {
		return err
	}
	// flag dependencies as enabled/disabled
	processDependencyTags(c.Metadata.Dependencies, cvals)
	processDependencyConditions(c.Metadata.Dependencies, cvals, path)
	// make a map of charts to remove
	rm := map[string]struct{}{}
	for _, r := range c.Metadata.Dependencies {
		if !r.Enabled {
			// remove disabled chart
			rm[r.Name] = struct{}{}
		}
	}
	// don't keep disabled charts in new slice
	cd := []*chart.Chart{}
	copy(cd, c.Dependencies()[:0])
	for _, n := range c.Dependencies() {
		if _, ok := rm[n.Metadata.Name]; !ok {
			cd = append(cd, n)
		}
	}
	// don't keep disabled charts in metadata
	cdMetadata := []*chart.Dependency{}
	copy(cdMetadata, c.Metadata.Dependencies[:0])
	for _, n := range c.Metadata.Dependencies {
		if _, ok := rm[n.Name]; !ok {
			cdMetadata = append(cdMetadata, n)
		}
	}

	// recursively call self to process sub dependencies
	for _, t := range cd {
		subpath := path + t.Metadata.Name + "."
		if err := processDependencyEnabled(t, cvals, subpath); err != nil {
			return err
		}
	}
	// set the correct dependencies in metadata
	c.Metadata.Dependencies = nil
	c.Metadata.Dependencies = append(c.Metadata.Dependencies, cdMetadata...)
	c.SetDependencies(cd...)

	return nil
}

// pathToMap creates a nested map given a YAML path in dot notation.
func pathToMap(path string, data map[string]interface{}) map[string]interface{} {
	if path == "." {
		return data
	}
	return set(parsePath(path), data)
}

func set(path []string, data map[string]interface{}) map[string]interface{} {
	if len(path) == 0 {
		return nil
	}
	cur := data
	for i := len(path) - 1; i >= 0; i-- {
		cur = map[string]interface{}{path[i]: cur}
	}
	return cur
}

// processImportValues merges values from child to parent based on the chart's dependencies' ImportValues field.
func processImportValues(c *chart.Chart, merge bool) error {
	if c.Metadata.Dependencies == nil {
		return nil
	}
	// combine chart values and empty config to get Values
	var cvals Values
	var err error
	if merge {
		cvals, err = MergeValues(c, nil)
	} else {
		cvals, err = CoalesceValues(c, nil)
	}
	if err != nil {
		return err
	}
	b := make(map[string]interface{})
	// import values from each dependency if specified in import-values
	for _, r := range c.Metadata.Dependencies {
		var outiv []interface{}
		for _, riv := range r.ImportValues {
			switch iv := riv.(type) {
			case map[string]interface{}:
				child := iv["child"].(string)
				parent := iv["parent"].(string)

				outiv = append(outiv, map[string]string{
					"child":  child,
					"parent": parent,
				})

				// get child table
				vv, err := cvals.Table(r.Name + "." + child)
				if err != nil {
					log.Printf("Warning: ImportValues missing table from chart %s: %v", r.Name, err)
					continue
				}
				// create value map from child to be merged into parent
				if merge {
					b = MergeTables(b, pathToMap(parent, vv.AsMap()))
				} else {
					b = CoalesceTables(b, pathToMap(parent, vv.AsMap()))
				}
			case string:
				child := "exports." + iv
				outiv = append(outiv, map[string]string{
					"child":  child,
					"parent": ".",
				})
				vm, err := cvals.Table(r.Name + "." + child)
				if err != nil {
					log.Printf("Warning: ImportValues missing table: %v", err)
					continue
				}
				if merge {
					b = MergeTables(b, vm.AsMap())
				} else {
					b = CoalesceTables(b, vm.AsMap())
				}
			}
		}
		r.ImportValues = outiv
	}

	// Imported values from a child to a parent chart have a lower priority than
	// the parents values. This enables parent charts to import a large section
	// from a child and then override select parts. This is why b is merged into
	// cvals in the code below and not the other way around.
	if merge {
		// deep copying the cvals as there are cases where pointers can end
		// up in the cvals when they are copied onto b in ways that break things.
		cvals = deepCopyMap(cvals)
		c.Values = MergeTables(cvals, b)
	} else {
		// Trimming the nil values from cvals is needed for backwards compatibility.
		// Previously, the b value had been populated with cvals along with some
		// overrides. This caused the coalescing functionality to remove the
		// nil/null values. This trimming is for backwards compat.
		cvals = trimNilValues(cvals)
		c.Values = CoalesceTables(cvals, b)
	}

	return nil
}

func deepCopyMap(vals map[string]interface{}) map[string]interface{} {
	valsCopy, err := copystructure.Copy(vals)
	if err != nil {
		return vals
	}
	return valsCopy.(map[string]interface{})
}

func trimNilValues(vals map[string]interface{}) map[string]interface{} {
	valsCopy, err := copystructure.Copy(vals)
	if err != nil {
		return vals
	}
	valsCopyMap := valsCopy.(map[string]interface{})
	for key, val := range valsCopyMap {
		if val == nil {
			// Iterate over the values and remove nil keys
			delete(valsCopyMap, key)
		} else if istable(val) {
			// Recursively call into ourselves to remove keys from inner tables
			valsCopyMap[key] = trimNilValues(val.(map[string]interface{}))
		}
	}

	return valsCopyMap
}

// processDependencyImportValues imports specified chart values from child to parent.
func processDependencyImportValues(c *chart.Chart, merge bool) error {
	for _, d := range c.Dependencies() {
		// recurse
		if err := processDependencyImportValues(d, merge); err != nil {
			return err
		}
	}
	return processImportValues(c, merge)
}
