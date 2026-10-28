package devtool

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvParsingNeverExecutesShell(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	raw := "# comment\nexport TOKEN='$(never-execute)'\nURL=hello # comment\nSPACES=\"two words\"\nEMPTY=\n"
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	env, err := ReadEnv(path)
	if err != nil || env["TOKEN"] != "$(never-execute)" || env["URL"] != "hello" || env["SPACES"] != "two words" || env["EMPTY"] != "" {
		t.Fatalf("parse: %+v %v", env, err)
	}
	for _, raw := range []string{"BAD KEY=value", "1INVALID=value", "NO_EQUALS", "TOKEN='unclosed"} {
		_ = os.WriteFile(path, []byte(raw), 0600)
		if _, err = ReadEnv(path); err == nil {
			t.Fatal("invalid env accepted")
		}
	}
	if env, err = ReadEnv(path + "missing"); err != nil || len(env) != 0 {
		t.Fatal(err)
	}
	merged := mergeEnv([]string{"A=old", "KEEP=value"}, map[string]string{"A": "new"})
	if strings.Contains(strings.Join(merged, "\n"), "A=old") {
		t.Fatal("duplicate env keys")
	}
}
func TestRootTargetsAndStartupFailure(t *testing.T) {
	root, err := RepoRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	targets, err := Targets(root)
	if err != nil || len(targets) != 11 {
		t.Fatalf("workspace modules missing: %+v %v", targets, err)
	}
	if _, err = RepoRoot(t.TempDir()); err == nil {
		t.Fatal("non-repository accepted")
	}
	if _, err = Targets(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing directory accepted")
	}
	if _, err = Start(context.Background(), Options{Mode: "unknown"}); err == nil {
		t.Fatal("invalid mode accepted")
	}
	cluster, err := Start(context.Background(), Options{Root: root, BinDir: filepath.Join(t.TempDir(), "missing"), DataDir: t.TempDir(), Mode: "demo", RandomPorts: true, Output: io.Discard})
	if err == nil {
		cluster.Close()
		t.Fatal("missing binary accepted")
	}
}
