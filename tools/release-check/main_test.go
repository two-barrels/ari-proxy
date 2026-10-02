package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnapshotPackagesSourcesWithoutSiblingReplacement(t *testing.T) {
	root, proxy := t.TempDir(), t.TempDir()
	mod := "module github.com/two-barrels/ari-proxy/v6\n\ngo 1.25.0\n\nrequire github.com/two-barrels/ari/v6 v6.0.0\nreplace github.com/two-barrels/ari/v6 => ../ari\n"
	for name, data := range map[string]string{"go.mod": mod, "new.go": "package example\n", ".git/private.go": "secret", "credential.txt": "secret"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := snapshot(root, proxy, "github.com/two-barrels/ari-proxy/v6", "v6.0.0-dev.test", "v6.0.0-dev.test", environment(map[string]string{"GOWORK": "off"})); err != nil {
		t.Fatal(err)
	}
	archives, err := filepath.Glob(filepath.Join(proxy, "github.com", "two-barrels", "ari-proxy", "v6", "@v", "*.zip"))
	if err != nil || len(archives) != 1 {
		t.Fatalf("archives=%v, err=%v", archives, err)
	}
	archive, err := zip.OpenReader(archives[0])
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	foundSource, foundMod := false, false
	for _, file := range archive.File {
		if strings.Contains(file.Name, "private") || strings.Contains(file.Name, "credential") {
			t.Fatalf("unexpected file: %s", file.Name)
		}
		if strings.HasSuffix(file.Name, "/new.go") {
			foundSource = true
		}
		if strings.HasSuffix(file.Name, "/go.mod") {
			reader, err := file.Open()
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), "replace") || !strings.Contains(string(data), "v6.0.0-dev.test") {
				t.Fatalf("invalid archived module: %s", data)
			}
			foundMod = true
		}
	}
	if !foundSource || !foundMod {
		t.Fatal("missing source or module file")
	}
	original, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if string(original) != mod {
		t.Fatal("modified original module")
	}
}

func TestTagModeRequiresBothVersions(t *testing.T) {
	if err := run("", "", "v6.0.0", ""); err == nil {
		t.Fatal("accepted an incomplete tag check")
	}
}
