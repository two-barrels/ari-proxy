package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/two-barrels/ari/v6/testfixtures"
)

type asteriskVersion struct{ major, minor, patch int }

func parseAsteriskVersion(value string) (asteriskVersion, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return asteriskVersion{}, fmt.Errorf("invalid Asterisk version %q", value)
	}
	var numbers [3]int
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return asteriskVersion{}, fmt.Errorf("invalid Asterisk version %q", value)
		}
		numbers[i] = number
	}
	return asteriskVersion{numbers[0], numbers[1], numbers[2]}, nil
}

// The Swagger "since" list records separate backport boundaries for supported
// release lines. A boundary on the requested major takes precedence over one
// on an older major, as with bridge variables in 20.20, 22.10, and 23.4.
func supportsAtVersion(since []string, target asteriskVersion) (bool, error) {
	var sameMajor *asteriskVersion
	olderMajor := false
	for _, value := range since {
		introduced, err := parseAsteriskVersion(value)
		if err != nil {
			return false, err
		}
		if introduced.major == target.major {
			if sameMajor == nil || introduced.minor < sameMajor.minor ||
				(introduced.minor == sameMajor.minor && introduced.patch < sameMajor.patch) {
				copy := introduced
				sameMajor = &copy
			}
		} else if introduced.major < target.major {
			olderMajor = true
		}
	}
	if sameMajor != nil {
		return target.minor > sameMajor.minor ||
			(target.minor == sameMajor.minor && target.patch >= sameMajor.patch), nil
	}
	return olderMajor, nil
}

func TestAsterisk20_22_23VersionBoundaries(t *testing.T) {
	spec, err := readManifest(filepath.Join("..", "..", "docs", "ari-23-spec-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range testfixtures.Asterisk20_22_23 {
		t.Run(fixture.Method+fixture.Path+"@"+fixture.Version, func(t *testing.T) {
			var since []string
			for _, op := range spec.Operations {
				if op.Method == fixture.Method && op.Path == fixture.Path {
					since = op.Since
					break
				}
			}
			if len(since) == 0 {
				t.Fatalf("missing versioned operation %s %s", fixture.Method, fixture.Path)
			}
			version, err := parseAsteriskVersion(fixture.Version)
			if err != nil {
				t.Fatal(err)
			}
			available, err := supportsAtVersion(since, version)
			if err != nil {
				t.Fatal(err)
			}
			if available != fixture.Available {
				t.Errorf("%s %s at %s: available=%t, want %t (since=%v)",
					fixture.Method, fixture.Path, fixture.Version, available, fixture.Available, since)
			}
		})
	}
}
