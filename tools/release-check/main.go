// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

// release-check builds an external consumer from packaged snapshots or tags.
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func command(dir string, env []string, args ...string) ([]byte, error) {
	c := exec.Command("go", args...)
	c.Dir, c.Env = dir, env
	out, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return out, nil
}

func modulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("missing module declaration in %s", root)
}

// Package build inputs, including uncommitted sources, without Git metadata or
// local replacements. Each archive is consumed through the Go module proxy.
func snapshot(root, proxyDir, path, version, ariVersion string, env []string) error {
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return err
	}
	dir := filepath.Join(proxyDir, path, "@v")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	modfile := filepath.Join(dir, version+".mod")
	if err := os.WriteFile(modfile, mod, 0600); err != nil {
		return err
	}
	if ariVersion != "" {
		if _, err := command(root, env, "mod", "edit", "-modfile="+modfile, "-dropreplace=github.com/two-barrels/ari/v6", "-require=github.com/two-barrels/ari/v6@"+ariVersion); err != nil {
			return err
		}
	}
	mod, err = os.ReadFile(modfile)
	if err != nil {
		return err
	}
	info, _ := json.Marshal(map[string]any{"Version": version, "Time": time.Now().UTC()})
	if err := os.WriteFile(filepath.Join(dir, version+".info"), info, 0600); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(dir, version+".zip"))
	if err != nil {
		return err
	}
	defer file.Close()
	archive := zip.NewWriter(file)
	err = filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		// The repositories have no embedded assets; retain source and fixture inputs.
		switch filepath.Ext(rel) {
		case ".go", ".mod", ".sum", ".json", ".tmpl", ".csv", ".notice":
		default:
			if rel != "LICENSE" && rel != "NOTICE" {
				return nil
			}
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if rel == "go.mod" {
			data = mod
		}
		writer, err := archive.Create(path + "@" + version + "/" + filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		return err
	})
	closeErr := archive.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func environment(overrides map[string]string) []string {
	var env []string
	for _, item := range os.Environ() {
		key := strings.SplitN(item, "=", 2)[0]
		if _, replaced := overrides[key]; !replaced {
			env = append(env, item)
		}
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}

func run(ariRoot, proxyRoot, ariVersion, proxyVersion string) error {
	tagged := ariVersion != "" || proxyVersion != ""
	if tagged && (ariVersion == "" || proxyVersion == "") {
		return fmt.Errorf("tag mode requires both --ari-version and --proxy-version")
	}
	temp, err := os.MkdirTemp("", "ari-release-check-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	overrides := map[string]string{"GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "auto"}
	env := environment(overrides)
	proxyPath := "github.com/two-barrels/ari-proxy/v6"
	if !tagged {
		ariRoot, err = filepath.Abs(ariRoot)
		if err != nil {
			return err
		}
		proxyRoot, err = filepath.Abs(proxyRoot)
		if err != nil {
			return err
		}
		ariPath, err := modulePath(ariRoot)
		if err != nil {
			return err
		}
		if ariPath != "github.com/two-barrels/ari/v6" {
			return fmt.Errorf("unexpected ARI module %s", ariPath)
		}
		proxyPath, err = modulePath(proxyRoot)
		if err != nil {
			return err
		}
		if proxyPath != "github.com/two-barrels/ari-proxy/v6" {
			return fmt.Errorf("unexpected proxy module %s", proxyPath)
		}
		suffix := fmt.Sprintf("%d", time.Now().UnixNano())
		ariVersion = "v6.0.0-dev." + suffix
		proxyVersion = "v6.0.0-dev." + suffix
		localProxy := filepath.Join(temp, "module-proxy")
		if err := snapshot(ariRoot, localProxy, ariPath, ariVersion, "", env); err != nil {
			return err
		}
		if err := snapshot(proxyRoot, localProxy, proxyPath, proxyVersion, ariVersion, env); err != nil {
			return err
		}
		upstream, err := command(temp, env, "env", "GOPROXY")
		if err != nil {
			return err
		}
		overrides["GOPROXY"] = "file://" + filepath.ToSlash(localProxy) + "," + strings.TrimSpace(string(upstream))
		overrides["GONOPROXY"] = "none"
		overrides["GONOSUMDB"] = "github.com/two-barrels/ari/v6," + proxyPath
		// Snapshot versions must not leave synthetic entries in the user's cache.
		cache, err := command(temp, env, "env", "GOMODCACHE")
		if err != nil {
			return err
		}
		overrides["GOMODCACHE"] = filepath.Join(temp, "cache")
		// Reuse already-downloaded dependencies via the existing cache's proxy files.
		overrides["GOPROXY"] = "file://" + filepath.ToSlash(localProxy) + ",file://" + filepath.ToSlash(filepath.Join(strings.TrimSpace(string(cache)), "cache", "download")) + "," + strings.TrimSpace(string(upstream))
		env = environment(overrides)
	}
	consumer := filepath.Join(temp, "consumer")
	if err := os.Mkdir(consumer, 0700); err != nil {
		return err
	}
	mod := fmt.Sprintf("module release-consumer\n\ngo 1.25.0\n\nrequire (\n github.com/two-barrels/ari/v6 %s\n %s %s\n)\n", ariVersion, proxyPath, proxyVersion)
	if err := os.WriteFile(filepath.Join(consumer, "go.mod"), []byte(mod), 0600); err != nil {
		return err
	}
	source := fmt.Sprintf(`package consumer
import (
 "context"
 "github.com/two-barrels/ari/v6"
 "github.com/two-barrels/ari/v6/client/native"
 "github.com/two-barrels/ari/v6/client/arimocks"
 "github.com/two-barrels/ari/v6/ext/play"
 "%s/client"
 "%s/server"
)
var _ ari.Client = native.New(nil)
var _ ari.Client = (*client.Client)(nil)
var _ ari.Channel = (*arimocks.Channel)(nil)
var _ = server.New
var _ = play.Prompt
func compileNewAPI(c ari.Client, ctx context.Context, key *ari.Key) {
 _, _ = c.Asterisk().InfoWithOptions(key, ari.AsteriskInfoOptions{})
 _, _ = c.Bridge().CreateWithoutID(key, ari.BridgeCreateOptions{})
 _, _ = c.Channel().PlayOnCollection(key, "", ari.ChannelPlayOptions{})
 file, err := c.StoredRecording().File(ctx, key)
 if err == nil { _ = file.Body.Close() }
}
`, proxyPath, proxyPath)
	if err := os.WriteFile(filepath.Join(consumer, "consumer.go"), []byte(source), 0600); err != nil {
		return err
	}
	if _, err := command(consumer, env, "mod", "tidy"); err != nil {
		return err
	}
	if _, err := command(consumer, env, "build", "-mod=mod", "./...", "github.com/two-barrels/ari/v6/...", proxyPath+"/..."); err != nil {
		return err
	}
	graph, err := command(consumer, env, "list", "-m", "-json", "all")
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(graph)))
	for decoder.More() {
		var module struct {
			Path    string
			Replace *json.RawMessage
		}
		if err := decoder.Decode(&module); err != nil {
			return err
		}
		if module.Replace != nil {
			return fmt.Errorf("consumer resolved a replacement for %s", module.Path)
		}
	}
	mode := "packaged snapshots"
	if tagged {
		mode = "published tags"
	}
	fmt.Printf("Standalone consumer OK (%s): ari %s, %s %s; no workspace or replacements\n", mode, ariVersion, proxyPath, proxyVersion)
	return nil
}

func main() {
	ariRoot := flag.String("ari-dir", "../ari", "ARI checkout for snapshot mode")
	proxyRoot := flag.String("proxy-dir", ".", "proxy checkout for snapshot mode")
	ariVersion := flag.String("ari-version", "", "published ARI v6 version")
	proxyVersion := flag.String("proxy-version", "", "published proxy v6 version")
	flag.Parse()
	if err := run(*ariRoot, *proxyRoot, *ariVersion, *proxyVersion); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
