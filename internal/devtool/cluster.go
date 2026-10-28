// Package devtool runs the same local service stack on Windows, macOS and Linux.
package devtool

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var Services = []string{"session-service", "ai-service", "map-service", "provider-service", "route-service", "travel-service", "gateway"}
var defaultPorts = map[string]int{"gateway": 8080, "travel-service": 50051, "session-service": 50052, "user-service": 50053, "order-service": 50054, "ai-service": 50055, "map-service": 50056, "route-service": 50057, "provider-service": 50058}

type Options struct {
	Root, BinDir, DataDir, Mode, EnvFile string
	RandomPorts                          bool
	ExtraEnv                             map[string]string
	Output                               io.Writer
}
type child struct {
	cmd  *exec.Cmd
	done chan struct{}
	err  error
	log  *os.File
}
type Cluster struct {
	URL      string
	Ports    map[string]int
	children []*child
	cancel   context.CancelFunc
}

func RepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			if _, err = os.Stat(filepath.Join(dir, "ABOUT.md")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("run this command inside the godnota repository")
		}
		dir = parent
	}
}

// ReadEnv parses configuration as data. It never sources a shell or expands commands.
func ReadEnv(path string) (map[string]string, error) {
	values := map[string]string{}
	if path == "" {
		return values, nil
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return values, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	line := 0
	for scan.Scan() {
		line++
		text := strings.TrimSpace(scan.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		text = strings.TrimPrefix(text, "export ")
		key, value, ok := strings.Cut(text, "=")
		key = strings.TrimSpace(key)
		if !ok || !envKey(key) {
			return nil, fmt.Errorf("invalid env entry at line %d", line)
		}
		value = strings.TrimSpace(value)
		if len(value) > 0 && (value[0] == '\'' || value[0] == '"') {
			if len(value) < 2 || value[len(value)-1] != value[0] {
				return nil, fmt.Errorf("unclosed env quote at line %d", line)
			}
			value = value[1 : len(value)-1]
		} else if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}
		values[key] = value
	}
	return values, scan.Err()
}
func envKey(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || i > 0 && r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
func mergeEnv(base []string, values map[string]string) []string {
	result := []string{}
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		found := false
		for k := range values {
			if k == key || runtime.GOOS == "windows" && strings.EqualFold(k, key) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, entry)
		}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}
func binaryName(service string) string {
	if runtime.GOOS == "windows" {
		return service + ".exe"
	}
	return service
}

func Build(ctx context.Context, root, bin string, out io.Writer) error {
	if err := os.MkdirAll(bin, 0755); err != nil {
		return err
	}
	for _, name := range Services {
		fmt.Fprintln(out, "build", name)
		cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(bin, binaryName(name)), "./"+name+"/cmd/app")
		cmd.Dir = root
		cmd.Env = mergeEnv(os.Environ(), map[string]string{"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "CGO_ENABLED": "0"})
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build %s: %w", name, err)
		}
	}
	return nil
}

func Start(ctx context.Context, opt Options) (cluster *Cluster, err error) {
	if opt.Mode != "demo" && opt.Mode != "live" {
		return nil, errors.New("mode must be demo or live")
	}
	if opt.Output == nil {
		opt.Output = io.Discard
	}
	if err = os.MkdirAll(filepath.Join(opt.DataDir, "logs"), 0700); err != nil {
		return nil, err
	}
	configured, err := ReadEnv(opt.EnvFile)
	if err != nil {
		return nil, err
	}
	for key := range configured {
		if value, exists := os.LookupEnv(key); exists {
			configured[key] = value
		}
	}
	for key, value := range opt.ExtraEnv {
		configured[key] = value
	}
	configured["PROVIDER_MODE"] = opt.Mode
	configured["AUTH_DATA_FILE"] = filepath.Join(opt.DataDir, "auth.json")
	configured["PROVIDER_CONFIG_FILE"] = filepath.Join(opt.Root, "configs", "providers.json")
	ports := map[string]int{}
	reserved := map[string]net.Listener{}
	defer func() {
		for _, l := range reserved {
			_ = l.Close()
		}
	}()
	for _, name := range Services {
		port := defaultPorts[name]
		if opt.RandomPorts {
			port = 0
		}
		l, e := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if e != nil {
			return nil, fmt.Errorf("%s port %d unavailable: %w", name, port, e)
		}
		reserved[name] = l
		ports[name] = l.Addr().(*net.TCPAddr).Port
	}
	for _, entry := range []struct{ key, name string }{{"SESSION_GRPC_ADDR", "session-service"}, {"TRAVEL_GRPC_ADDR", "travel-service"}, {"MAP_GRPC_ADDR", "map-service"}, {"ROUTE_GRPC_ADDR", "route-service"}, {"PROVIDER_GRPC_ADDR", "provider-service"}, {"AI_GRPC_ADDR", "ai-service"}} {
		configured[entry.key] = net.JoinHostPort("127.0.0.1", strconv.Itoa(ports[entry.name]))
	}
	for key, value := range map[string]string{"USER_GRPC_ADDR": "127.0.0.1:50053", "ORDER_GRPC_ADDR": "127.0.0.1:50054"} {
		configured[key] = value
	}
	childCtx, cancel := context.WithCancel(ctx)
	cluster = &Cluster{URL: "http://127.0.0.1:" + strconv.Itoa(ports["gateway"]), Ports: ports, cancel: cancel}
	defer func() {
		if err != nil {
			cluster.Close()
		}
	}()
	for _, name := range Services {
		_ = reserved[name].Close()
		delete(reserved, name)
		log, e := os.OpenFile(filepath.Join(opt.DataDir, "logs", name+".log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if e != nil {
			return cluster, e
		}
		values := map[string]string{}
		for key, value := range configured {
			values[key] = value
		}
		values["GRPC_PORT"] = strconv.Itoa(ports[name])
		values["HTTP_PORT"] = strconv.Itoa(ports[name])
		cmd := exec.CommandContext(childCtx, filepath.Join(opt.BinDir, binaryName(name)))
		cmd.Dir = opt.Root
		cmd.Env = mergeEnv(os.Environ(), values)
		cmd.Stdout = log
		cmd.Stderr = log
		if e = cmd.Start(); e != nil {
			_ = log.Close()
			return cluster, fmt.Errorf("start %s: %w", name, e)
		}
		process := &child{cmd: cmd, done: make(chan struct{}), log: log}
		cluster.children = append(cluster.children, process)
		go func() { process.err = cmd.Wait(); _ = log.Close(); close(process.done) }()
	}
	ready, cancelReady := context.WithTimeout(childCtx, 30*time.Second)
	defer cancelReady()
	if err = cluster.ready(ready); err != nil {
		return cluster, err
	}
	fmt.Fprintln(opt.Output, "API:", cluster.URL, "mode:", opt.Mode)
	return cluster, nil
}
func (c *Cluster) ready(ctx context.Context) error {
	client := &http.Client{Timeout: time.Second}
	timer := time.NewTicker(100 * time.Millisecond)
	defer timer.Stop()
	for {
		if err := c.check(); err != nil {
			return err
		}
		all := true
		for _, port := range c.Ports {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 100*time.Millisecond)
			if err != nil {
				all = false
				break
			}
			_ = conn.Close()
		}
		if all {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"/health", nil)
			response, err := client.Do(req)
			if err == nil {
				_ = response.Body.Close()
				if response.StatusCode == 200 {
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("services did not become ready: %w (see logs)", ctx.Err())
		case <-timer.C:
		}
	}
}
func (c *Cluster) check() error {
	for _, p := range c.children {
		select {
		case <-p.done:
			return fmt.Errorf("service %s exited: %v (see logs)", filepath.Base(p.cmd.Path), p.err)
		default:
		}
	}
	return nil
}
func (c *Cluster) Wait(ctx context.Context) error {
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	for {
		if err := c.check(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}
func (c *Cluster) Close() {
	if c == nil {
		return
	}
	c.cancel()
	for _, p := range c.children {
		<-p.done
	}
}

func Targets(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	targets := []string{"./..."}
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(root, entry.Name(), "go.mod")); err == nil {
				targets = append(targets, "./"+entry.Name()+"/...")
			}
		}
	}
	return targets, nil
}
