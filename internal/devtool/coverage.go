package devtool

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Scope is explicit: generated code and the team's unrelated stubs are excluded.
// Adding an implementation file to this part requires adding it here as well.
var coverageFiles = map[string]string{
	"github.com/ostrovok-hackathon-2026/godnota/providerkit/api.go":             "providers",
	"github.com/ostrovok-hackathon-2026/godnota/providerkit/parts.go":           "providers",
	"github.com/ostrovok-hackathon-2026/godnota/providerkit/normalize.go":       "providers",
	"github.com/godnota/pkg/tripconfig/config.go":                               "config bridge",
	"github.com/godnota/gateway/internal/auth/store.go":                         "auth store",
	"github.com/godnota/gateway/internal/httpapi/auth.go":                       "HTTP API",
	"github.com/godnota/gateway/internal/httpapi/server.go":                     "HTTP API",
	"github.com/godnota/gateway/internal/httpapi/cors.go":                       "HTTP API",
	"github.com/godnota/gateway/internal/adapter/adapter.go":                    "gateway gRPC adapter",
	"github.com/godnota/map-service/internal/controller/mapping/mapping.go":     "mapping RPC",
	"github.com/godnota/provider-service/internal/controller/provider/parts.go": "provider RPC",
	"github.com/godnota/session-service/internal/controller/session/session.go": "sessions",
	"github.com/godnota/session-service/internal/repository/session/memory.go":  "sessions",
	"github.com/godnota/session-service/internal/usecase/session/session.go":    "sessions",
	"github.com/godnota/travel-service/internal/controller/travel/planning.go":  "planning",
	"github.com/godnota/route-service/internal/controller/route/route.go":       "route handoff",
}

func CoveragePackages() string {
	set := map[string]bool{}
	for file := range coverageFiles {
		set[path.Dir(file)] = true
	}
	packages := []string{}
	for name := range set {
		packages = append(packages, name)
	}
	sort.Strings(packages)
	return strings.Join(packages, ",")
}

type covered struct{ hit, total int }

func CoverageReport(profile, output string, minimum float64) error {
	file, err := os.Open(profile)
	if err != nil {
		return err
	}
	defer file.Close()
	groups := map[string]covered{}
	type block struct {
		group string
		count int
		hit   bool
	}
	blocks := map[string]block{}
	seen := map[string]bool{}
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if len(fields) != 3 {
			continue
		}
		name, _, _ := strings.Cut(fields[0], ":")
		name = strings.ReplaceAll(name, "\\", "/")
		group, ok := coverageFiles[name]
		if !ok {
			continue
		}
		count, e := strconv.Atoi(fields[1])
		if e != nil {
			return e
		}
		hits, e := strconv.Atoi(fields[2])
		if e != nil {
			return e
		}
		old := blocks[fields[0]]
		blocks[fields[0]] = block{group: group, count: count, hit: old.hit || hits > 0}
		seen[name] = true
	}
	if err = scan.Err(); err != nil {
		return err
	}
	for _, block := range blocks {
		value := groups[block.group]
		value.total += block.count
		if block.hit {
			value.hit += block.count
		}
		groups[block.group] = value
	}
	for name := range coverageFiles {
		if !seen[name] {
			return fmt.Errorf("coverage missing for owned file %s", name)
		}
	}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	report := "# Coverage: owned implementation files\n\nUnit tests and in-process HTTP/gRPC integrations; external-process E2E is reported separately. Generated protobuf, launchers and unrelated stubs are not included.\n\n| Scope | Covered statements | Total | Coverage |\n|---|---:|---:|---:|\n"
	total := covered{}
	for _, key := range keys {
		v := groups[key]
		report += fmt.Sprintf("| %s | %d | %d | %.1f%% |\n", key, v.hit, v.total, 100*float64(v.hit)/float64(v.total))
		total.hit += v.hit
		total.total += v.total
	}
	percent := 100 * float64(total.hit) / float64(total.total)
	report += fmt.Sprintf("| **Total** | **%d** | **%d** | **%.1f%%** |\n\nRequired aggregate coverage: %.0f%%. This measures statements, not complete behavioral coverage.\n", total.hit, total.total, percent, minimum)
	if err = os.WriteFile(output, []byte(report), 0644); err != nil {
		return err
	}
	fmt.Print(report)
	if percent < minimum {
		return fmt.Errorf("owned code coverage %.1f%% is below %.0f%%", percent, minimum)
	}
	return nil
}
