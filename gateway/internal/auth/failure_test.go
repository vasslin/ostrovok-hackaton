package auth

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInvalidInputsAndStorageFiles(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("empty path accepted")
	}
	dir := t.TempDir()
	for _, data := range []string{"broken json", "{}", "null"} {
		path := filepath.Join(dir, "invalid.json")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Open(path); err == nil {
			t.Fatal("invalid storage accepted", data)
		}
	}
	if _, err := Open(dir); err == nil {
		t.Fatal("directory accepted as file")
	}
	s, _ := Open(filepath.Join(dir, "valid.json"))
	for _, c := range []Credentials{{Email: "bad", Password: "long-enough-password"}, {Email: "Name <a@example.com>", Password: "long-enough-password"}, {Email: "a@example.com", Password: "short"}, {Email: "a@example.com", Password: strings.Repeat("x", 129)}} {
		if _, err := s.Register(c); err != ErrInput {
			t.Fatal("invalid credentials accepted", err)
		}
	}
	if _, err := s.Login(Credentials{Email: "missing@example.com", Password: "long-enough-password"}); err != ErrCredentials {
		t.Fatal(err)
	}
	if _, err := s.Login(Credentials{Password: strings.Repeat("x", 129)}); err != ErrCredentials {
		t.Fatal(err)
	}
}
func TestStorageFailuresRollbackAuthenticationState(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(filepath.Join(dir, "auth.json"))
	credentials := Credentials{Email: "rollback@example.com", Password: "long-demo-password"}
	login, err := s.Register(credentials)
	if err != nil {
		t.Fatal(err)
	}
	goodPath := s.path
	blocker := filepath.Join(dir, "not-a-directory")
	if err = os.WriteFile(blocker, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	s.path = filepath.Join(blocker, "auth.json")
	if _, err = s.Login(credentials); err == nil {
		t.Fatal("unpersisted login succeeded")
	}
	if _, err = s.Authenticate(login.AccessToken); err != nil {
		t.Fatal("failed login revoked old token", err)
	}
	if err = s.Logout(login.AccessToken); err == nil {
		t.Fatal("unpersisted logout succeeded")
	}
	if _, err = s.Authenticate(login.AccessToken); err != nil {
		t.Fatal("failed logout revoked token", err)
	}
	failed := Credentials{Email: "other@example.com", Password: "long-demo-password"}
	if _, err = s.Register(failed); err == nil {
		t.Fatal("unpersisted registration succeeded")
	}
	s.path = goodPath
	if _, err = s.Register(failed); err != nil {
		t.Fatal("failed registration left account in memory", err)
	}
	s.now = func() time.Time { return time.Now().Add(25 * time.Hour) }
	if _, err = s.Login(credentials); err != nil {
		t.Fatal(err)
	}
	for _, v := range s.data.Sessions {
		if !s.now().Before(v.ExpiresAt) {
			t.Fatal("expired session retained")
		}
	}
}
func TestConcurrentDuplicateRegistrations(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "auth.json"))
	var wg sync.WaitGroup
	results := make(chan error, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Register(Credentials{Email: "same@example.com", Password: "long-demo-password"})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if err != ErrExists {
			t.Error(err)
		}
	}
	if success != 1 {
		t.Fatalf("duplicate accounts created: %d", success)
	}
}
