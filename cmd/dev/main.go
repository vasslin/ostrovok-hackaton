// dev provides shell-independent commands for all supported development OSes.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ostrovok-hackathon-2026/godnota/internal/devtool"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: go run ./cmd/dev run | test [-race] [-cover] | e2e | smoke [-url URL]")
	}
	root, err := devtool.RepoRoot(".")
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	flags := flag.NewFlagSet(os.Args[1], flag.ContinueOnError)
	switch os.Args[1] {
	case "run":
		mode := flags.String("mode", "demo", "demo or live")
		random := flags.Bool("random-ports", false, "use free ports instead of default ports")
		env := flags.String("env-file", ".env", "local configuration file")
		if err = flags.Parse(os.Args[2:]); err != nil {
			return err
		}
		bin := filepath.Join(root, "bin")
		if err = devtool.Build(ctx, root, bin, os.Stdout); err != nil {
			return err
		}
		envPath := *env
		if !filepath.IsAbs(envPath) {
			envPath = filepath.Join(root, envPath)
		}
		cluster, err := devtool.Start(ctx, devtool.Options{Root: root, BinDir: bin, DataDir: filepath.Join(root, "data"), Mode: *mode, EnvFile: envPath, RandomPorts: *random, Output: os.Stdout})
		if err != nil {
			return err
		}
		defer cluster.Close()
		fmt.Println("Docs:", cluster.URL+"/docs", "; Ctrl+C stops all child services")
		return cluster.Wait(ctx)
	case "test":
		race := flags.Bool("race", false, "enable Go race detector (requires C compiler)")
		coverage := flags.Bool("cover", false, "write coverage profiles and HTML report")
		if err = flags.Parse(os.Args[2:]); err != nil {
			return err
		}
		targets, err := devtool.Targets(root)
		if err != nil {
			return err
		}
		args := []string{"test", "-timeout=3m"}
		if *race {
			args = append(args, "-race")
		}
		if *coverage {
			if err = os.MkdirAll(filepath.Join(root, "reports"), 0755); err != nil {
				return err
			}
			args = append(args, "-covermode=atomic", "-coverprofile=reports/unit.coverprofile", "-coverpkg="+devtool.CoveragePackages())
		}
		args = append(args, targets...)
		if err = command(ctx, root, args...); err != nil {
			return err
		}
		if err = command(ctx, root, append([]string{"vet"}, targets...)...); err != nil {
			return err
		}
		if *coverage {
			if err = command(ctx, root, "tool", "cover", "-html=reports/unit.coverprofile", "-o=reports/coverage.html"); err != nil {
				return err
			}
			return devtool.CoverageReport(filepath.Join(root, "reports/unit.coverprofile"), filepath.Join(root, "reports/coverage.md"), 80)
		}
		return nil
	case "e2e":
		return command(ctx, root, "test", "-tags=e2e", "-count=1", "-v", "-timeout=5m", "./e2e")
	case "smoke":
		url := flags.String("url", "http://127.0.0.1:8080", "running gateway URL")
		if err = flags.Parse(os.Args[2:]); err != nil {
			return err
		}
		return devtool.Smoke(ctx, *url, os.Stdout)
	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}
func command(ctx context.Context, root string, args ...string) error {
	fmt.Println("go", args)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
