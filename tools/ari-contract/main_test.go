package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func loadCheckedInContract(t *testing.T) (manifest, []row, []row) {
	t.Helper()
	root := filepath.Join("..", "..")
	spec, err := readManifest(filepath.Join(root, "docs/ari-23-spec-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	routes, err := readCSV(filepath.Join(root, "docs/ari-23-endpoint-inventory.csv"))
	if err != nil {
		t.Fatal(err)
	}
	parameters, err := readCSV(filepath.Join(root, "docs/ari-23-parameter-coverage.csv"))
	if err != nil {
		t.Fatal(err)
	}
	return spec, routes, parameters
}

func TestCheckedInContract(t *testing.T) {
	spec, routes, parameters := loadCheckedInContract(t)
	if problems := checkContract(filepath.Join("..", ".."), spec, routes, parameters); len(problems) > 0 {
		t.Fatal(strings.Join(problems, "\n"))
	}
	verified := 0
	for _, r := range parameters {
		if r["ari_wire_test"] != "" && r["proxy_wire_test"] != "" {
			verified++
		}
	}
	if verified < 175 {
		t.Fatalf("verified %d parameters, want at least 175", verified)
	}
}

func TestNewOperationRequiresClassification(t *testing.T) {
	spec, routes, parameters := loadCheckedInContract(t)
	spec.Operations = append(spec.Operations, operation{
		Resource: "channels", Method: "POST", Path: "/channels/newFeature", Name: "newFeature",
		PathParameters: []parameter{}, Parameters: []parameter{},
	})
	problems := checkContract(filepath.Join("..", ".."), spec, routes, parameters)
	if !containsProblem(problems, "unclassified operation: channels POST /channels/newFeature") {
		t.Fatal(problems)
	}
}

func TestNewParameterRequiresClassification(t *testing.T) {
	spec, routes, parameters := loadCheckedInContract(t)
	index := slices.IndexFunc(spec.Operations, func(op operation) bool { return op.Path == "/channels/create" })
	if index < 0 {
		t.Fatal("/channels/create absent from manifest")
	}
	spec.Operations[index].Parameters = append(spec.Operations[index].Parameters, parameter{Name: "newOption", Location: "query", Type: "string"})
	problems := checkContract(filepath.Join("..", ".."), spec, routes, parameters)
	if !containsProblem(problems, "unclassified parameter: channels POST /channels/create newOption") {
		t.Fatal(problems)
	}
}

func TestChangedParameterReportsDrift(t *testing.T) {
	spec, _, _ := loadCheckedInContract(t)
	current := spec
	current.Operations = slices.Clone(spec.Operations)
	current.Operations[0].Parameters = []parameter{{Name: "newOption", Location: "query", Type: "string"}}
	if !strings.Contains(describeDrift(spec, current), "changed operation or parameter") {
		t.Fatal(describeDrift(spec, current))
	}
}

func containsProblem(problems []string, want string) bool {
	return slices.ContainsFunc(problems, func(problem string) bool { return strings.Contains(problem, want) })
}
