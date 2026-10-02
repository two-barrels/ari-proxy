// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

// ari-contract checks the pinned Asterisk ARI specification against coverage inventories.
package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

var resourceFiles = []string{
	"applications", "asterisk", "bridges", "channels", "deviceStates", "endpoints",
	"events", "mailboxes", "playbacks", "recordings", "sounds",
}

type parameter struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  any    `json:"default,omitempty"`
}

type operation struct {
	Resource       string      `json:"resource"`
	Method         string      `json:"method"`
	Path           string      `json:"path"`
	Name           string      `json:"operation"`
	Response       string      `json:"response"`
	Since          []string    `json:"since"`
	PathParameters []parameter `json:"path_parameters"`
	Parameters     []parameter `json:"parameters"`
}

type manifest struct {
	Source        string      `json:"source"`
	Ref           string      `json:"ref"`
	ResourceFiles []string    `json:"resource_files"`
	Operations    []operation `json:"operations"`
}

type row map[string]string

func readManifest(path string) (manifest, error) {
	var result manifest
	data, err := os.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(data, &result)
	}
	return result, err
}

func readCSV(path string) ([]row, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	var rows []row
	for {
		values, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		item := make(row, len(headers))
		for i, header := range headers {
			item[header] = values[i]
		}
		rows = append(rows, item)
	}
	return rows, nil
}

func extractSpec(directory, ref string) (manifest, error) {
	result := manifest{Source: "https://github.com/asterisk/asterisk", Ref: ref, ResourceFiles: slices.Clone(resourceFiles), Operations: []operation{}}
	for _, resource := range resourceFiles {
		data, err := os.ReadFile(filepath.Join(directory, resource+".json"))
		if err != nil {
			return manifest{}, err
		}
		var spec struct {
			APIs []struct {
				Path       string `json:"path"`
				Operations []struct {
					Method     string   `json:"httpMethod"`
					Nickname   string   `json:"nickname"`
					Response   string   `json:"responseClass"`
					Since      []string `json:"since"`
					Parameters []struct {
						Name     string `json:"name"`
						Location string `json:"paramType"`
						Type     string `json:"dataType"`
						Required bool   `json:"required"`
						Default  any    `json:"defaultValue"`
					} `json:"parameters"`
				} `json:"operations"`
			} `json:"apis"`
		}
		if err := json.Unmarshal(data, &spec); err != nil {
			return manifest{}, fmt.Errorf("%s: %w", resource, err)
		}
		for _, api := range spec.APIs {
			for _, source := range api.Operations {
				item := operation{Resource: resource, Method: source.Method, Path: api.Path, Name: source.Nickname,
					Response: source.Response, Since: source.Since, PathParameters: []parameter{}, Parameters: []parameter{}}
				if item.Since == nil {
					item.Since = []string{}
				}
				for _, p := range source.Parameters {
					value := parameter{Name: p.Name, Location: p.Location, Type: p.Type, Required: p.Required, Default: p.Default}
					if p.Location == "path" {
						item.PathParameters = append(item.PathParameters, value)
					} else {
						item.Parameters = append(item.Parameters, value)
					}
				}
				result.Operations = append(result.Operations, item)
			}
		}
	}
	slices.SortFunc(result.Operations, func(a, b operation) int {
		return strings.Compare(routeKey(a.Resource, a.Path, a.Method), routeKey(b.Resource, b.Path, b.Method))
	})
	return result, nil
}

func routeKey(resource, path, method string) string { return resource + " " + method + " " + path }
func operationKey(op operation) string              { return routeKey(op.Resource, op.Path, op.Method) }
func rowKey(r row) string                           { return routeKey(r["resource"], r["path"], r["method"]) }
func parameterKey(route, name string) string        { return route + " " + name }

func checkSource(root, source, label string) []string {
	if source == "" {
		return []string{"missing implementation source for " + label}
	}
	var problems []string
	for _, path := range strings.Split(source, "|") {
		if !strings.HasSuffix(path, ".go") {
			problems = append(problems, fmt.Sprintf("invalid source path %q for %s", path, label))
			continue
		}
		full := filepath.Join(root, path)
		// The sibling ari checkout is absent when CI checks this repo alone.
		if strings.HasPrefix(path, "../ari/") {
			if _, err := os.Stat(filepath.Join(root, "../ari")); os.IsNotExist(err) {
				continue
			}
		}
		if info, err := os.Stat(full); err != nil || info.IsDir() {
			problems = append(problems, fmt.Sprintf("source file %q does not exist for %s", path, label))
		}
	}
	return problems
}

func checkContract(root string, spec manifest, routeRows, parameterRows []row) []string {
	var problems []string
	specRoutes := map[string]operation{}
	routes := map[string]row{}
	parameters := map[string]row{}
	for _, op := range spec.Operations {
		key := operationKey(op)
		if _, ok := specRoutes[key]; ok {
			problems = append(problems, "duplicate spec operation: "+key)
		}
		specRoutes[key] = op
	}
	for _, r := range routeRows {
		key := rowKey(r)
		if _, ok := routes[key]; ok {
			problems = append(problems, "duplicate coverage route: "+key)
		}
		routes[key] = r
	}
	for _, r := range parameterRows {
		key := parameterKey(rowKey(r), r["parameter"])
		if _, ok := parameters[key]; ok {
			problems = append(problems, "duplicate coverage parameter: "+key)
		}
		parameters[key] = r
	}
	for key := range specRoutes {
		if _, ok := routes[key]; !ok {
			problems = append(problems, "unclassified operation: "+key)
		}
	}
	for key := range routes {
		if _, ok := specRoutes[key]; !ok {
			problems = append(problems, "obsolete coverage operation: "+key)
		}
	}
	expected := map[string]parameter{}
	for key, op := range specRoutes {
		for _, p := range op.Parameters {
			expected[parameterKey(key, p.Name)] = p
		}
	}
	for key := range expected {
		if _, ok := parameters[key]; !ok {
			problems = append(problems, "unclassified parameter: "+key)
		}
	}
	for key := range parameters {
		if _, ok := expected[key]; !ok {
			problems = append(problems, "obsolete coverage parameter: "+key)
		}
	}
	for key, op := range specRoutes {
		r, ok := routes[key]
		if !ok {
			continue
		}
		if r["operation"] != op.Name {
			problems = append(problems, "operation name changed for "+key)
		}
		var display []string
		for _, p := range op.Parameters {
			name := p.Name
			if p.Required {
				name += " [required]"
			}
			display = append(display, name+" ("+p.Location+")")
		}
		if r["non_path_parameters"] != strings.Join(display, "; ") {
			problems = append(problems, "parameter list changed for "+key)
		}
		for _, library := range []string{"ari", "proxy"} {
			status, source, wire := r[library+"_route"], r[library+"_source"], r[library+"_wire_test"]
			switch status {
			case "missing":
				if source != "" {
					problems = append(problems, "missing "+library+" route has a source for "+key)
				}
			case "implemented", "alternative":
				problems = append(problems, checkSource(root, source, library+" route "+key)...)
			default:
				problems = append(problems, "unclassified "+library+" route status for "+key)
			}
			if wire != "" {
				if status == "missing" {
					problems = append(problems, "missing "+library+" route has a wire test for "+key)
				}
				problems = append(problems, checkSource(root, wire, library+" route test "+key)...)
			}
		}
		valid := map[string]bool{}
		for _, p := range op.Parameters {
			valid[p.Name] = true
		}
		for _, gap := range splitGaps(r["known_option_gaps"]) {
			if !valid[gap] {
				problems = append(problems, "unknown option gap "+gap+" for "+key)
				continue
			}
			p := parameters[parameterKey(key, gap)]
			if p != nil && p["ari_status"] != "missing" && p["proxy_status"] != "missing" {
				problems = append(problems, "stale option gap "+gap+" for "+key)
			}
		}
	}
	for key, p := range expected {
		r, ok := parameters[key]
		if !ok {
			continue
		}
		for column, value := range map[string]string{
			"location": p.Location, "type": p.Type, "required": fmt.Sprint(p.Required), "default": defaultString(p.Default),
		} {
			if column == "required" {
				value = strings.ToLower(value)
			}
			if r[column] != value {
				problems = append(problems, fmt.Sprintf("%s changed for %s: %q != %q", column, key, r[column], value))
			}
		}
		route := routes[rowKey(r)]
		for _, library := range []string{"ari", "proxy"} {
			status, source, wire := r[library+"_status"], r[library+"_source"], r[library+"_wire_test"]
			switch status {
			case "mapped":
				problems = append(problems, checkSource(root, source, library+" parameter "+key)...)
				if route != nil && route[library+"_route"] == "missing" {
					problems = append(problems, library+" parameter mapped on missing route "+key)
				}
			case "missing":
				if source != "" {
					problems = append(problems, "missing "+library+" parameter has a source for "+key)
				}
			default:
				problems = append(problems, "unclassified "+library+" parameter status for "+key)
			}
			if wire != "" {
				if status != "mapped" {
					problems = append(problems, "unmapped "+library+" parameter has a wire test for "+key)
				}
				problems = append(problems, checkSource(root, wire, library+" parameter test "+key)...)
			}
			if route != nil && status == "missing" && route[library+"_route"] != "missing" && !slices.Contains(splitGaps(route["known_option_gaps"]), r["parameter"]) {
				problems = append(problems, library+" option gap absent from route inventory: "+key)
			}
		}
	}
	slices.Sort(problems)
	return problems
}

func splitGaps(value string) []string {
	var gaps []string
	for _, part := range strings.Split(value, ";") {
		if name := strings.TrimSpace(part); name != "" {
			gaps = append(gaps, name)
		}
	}
	return gaps
}

func defaultString(value any) string {
	if value == nil {
		return ""
	}
	if b, ok := value.(bool); ok {
		if b {
			return "True"
		}
		return "False"
	}
	return fmt.Sprint(value)
}

func run(root, specDir, writeManifest, ref string) error {
	if writeManifest != "" {
		if specDir == "" || ref == "" {
			return errors.New("--write-manifest requires --spec-dir and --ref")
		}
		spec, err := extractSpec(specDir, ref)
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(writeManifest, append(data, '\n'), 0644); err != nil {
			return err
		}
		fmt.Printf("wrote %d operations to %s\n", len(spec.Operations), writeManifest)
		return nil
	}
	spec, err := readManifest(filepath.Join(root, "docs/ari-23-spec-manifest.json"))
	if err != nil {
		return err
	}
	if specDir != "" {
		current, err := extractSpec(specDir, spec.Ref)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(current, spec) {
			return fmt.Errorf("Asterisk specification differs from the pinned manifest:\n%s", describeDrift(spec, current))
		}
	}
	routes, err := readCSV(filepath.Join(root, "docs/ari-23-endpoint-inventory.csv"))
	if err != nil {
		return err
	}
	parameters, err := readCSV(filepath.Join(root, "docs/ari-23-parameter-coverage.csv"))
	if err != nil {
		return err
	}
	if problems := checkContract(root, spec, routes, parameters); len(problems) > 0 {
		return errors.New(strings.Join(problems, "\n"))
	}
	verified, total := 0, 0
	for _, r := range parameters {
		if r["ari_wire_test"] != "" && r["proxy_wire_test"] != "" {
			verified++
		}
	}
	for _, op := range spec.Operations {
		total += len(op.Parameters)
	}
	fmt.Printf("ARI contract OK: %d operations, %d parameters classified, %d verified on both wires\n", len(spec.Operations), total, verified)
	return nil
}

func describeDrift(old, current manifest) string {
	previous, next := map[string]operation{}, map[string]operation{}
	for _, op := range old.Operations {
		previous[operationKey(op)] = op
	}
	for _, op := range current.Operations {
		next[operationKey(op)] = op
	}
	var changes []string
	for key, op := range next {
		before, ok := previous[key]
		if !ok {
			changes = append(changes, "added operation: "+key)
		} else if !reflect.DeepEqual(before, op) {
			changes = append(changes, "changed operation or parameter: "+key)
		}
	}
	for key := range previous {
		if _, ok := next[key]; !ok {
			changes = append(changes, "removed operation: "+key)
		}
	}
	slices.Sort(changes)
	if len(changes) == 0 {
		return "manifest metadata changed"
	}
	return strings.Join(changes, "\n")
}

func main() {
	root := flag.String("root", ".", "ari-proxy checkout root")
	specDir := flag.String("spec-dir", "", "local Asterisk rest-api/api-docs directory")
	writeManifest := flag.String("write-manifest", "", "write a normalized manifest from --spec-dir")
	ref := flag.String("ref", "", "commit SHA for --write-manifest")
	flag.Parse()
	if err := run(*root, *specDir, *writeManifest, *ref); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
