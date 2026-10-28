package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAccountLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	s, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	c := Credentials{Email: " User@example.com ", Password: "a-long-demo-password"}
	login, e := s.Register(c)
	if e != nil {
		t.Fatal(e)
	}
	if login.User.Email != "user@example.com" {
		t.Fatal("email not normalized")
	}
	if _, e = s.Register(c); e != ErrExists {
		t.Fatal("duplicate accepted")
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), c.Password) || strings.Contains(string(raw), login.AccessToken) {
		t.Fatal("plaintext credential persisted")
	}
	s, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(login.AccessToken); e != nil {
		t.Fatal("session not persisted")
	}
	if _, e = s.Login(Credentials{Email: c.Email, Password: "wrong"}); e != ErrCredentials {
		t.Fatal("wrong password accepted")
	}
	next, e := s.Login(c)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(login.AccessToken); e == nil {
		t.Fatal("old session not rotated")
	}
	if e = s.Logout(next.AccessToken); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(next.AccessToken); e == nil {
		t.Fatal("logged out token accepted")
	}
}
func TestExpiredAndForgedTokens(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "auth.json"))
	login, e := s.Register(Credentials{Email: "test@example.com", Password: "long-demo-password"})
	if e != nil {
		t.Fatal(e)
	}
	s.now = func() time.Time { return time.Now().Add(25 * time.Hour) }
	if _, e = s.Authenticate(login.AccessToken); e == nil {
		t.Fatal("expired token accepted")
	}
	if _, e = s.Authenticate(strings.Repeat("x", 43)); e == nil {
		t.Fatal("forged token accepted")
	}
}
