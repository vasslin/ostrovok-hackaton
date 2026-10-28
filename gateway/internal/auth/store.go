// Package auth provides local hackathon accounts and revocable opaque sessions.
package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var ErrCredentials = errors.New("invalid email or password")
var ErrExists = errors.New("account already exists")
var ErrInput = errors.New("valid email and password of 12 to 128 characters required")

const iterations = 600000

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
type account struct {
	User User   `json:"user"`
	Salt []byte `json:"salt"`
	Hash []byte `json:"hash"`
}
type session struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
type disk struct {
	Accounts map[string]account `json:"accounts"`
	Sessions map[string]session `json:"sessions"`
}
type Store struct {
	mu   sync.Mutex
	path string
	data disk
	now  func() time.Time
}
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Login struct {
	User        User      `json:"user"`
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, now: time.Now, data: disk{Accounts: map[string]account{}, Sessions: map[string]session{}}}
	if path == "" {
		return nil, errors.New("AUTH_DATA_FILE is required")
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var persisted disk
	if err = json.Unmarshal(raw, &persisted); err != nil {
		return nil, errors.New("invalid auth data file")
	}
	if persisted.Accounts == nil || persisted.Sessions == nil {
		return nil, errors.New("incomplete auth data file")
	}
	s.data = persisted
	return s, nil
}

func randomToken() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}
func digest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func derive(password string, salt []byte) []byte {
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
	if err != nil {
		panic(err)
	}
	return key
}

func (s *Store) Register(c Credentials) (Login, error) {
	email := strings.ToLower(strings.TrimSpace(c.Email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 || utf8.RuneCountInString(c.Password) < 12 || len(c.Password) > 128 {
		return Login{}, ErrInput
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return Login{}, err
	}
	a := account{User: User{ID: randomToken(), Email: email}, Salt: salt, Hash: derive(c.Password, salt)}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.Accounts[email]; exists {
		return Login{}, ErrExists
	}
	s.data.Accounts[email] = a
	login, key := s.issue(a.User)
	if err := s.save(); err != nil {
		delete(s.data.Accounts, email)
		delete(s.data.Sessions, key)
		return Login{}, err
	}
	return login, nil
}

func (s *Store) Login(c Credentials) (Login, error) {
	email := strings.ToLower(strings.TrimSpace(c.Email))
	if len(c.Password) > 128 {
		return Login{}, ErrCredentials
	}
	s.mu.Lock()
	a, exists := s.data.Accounts[email]
	s.mu.Unlock()
	if !exists {
		a.Salt = make([]byte, 16)
		a.Hash = make([]byte, 32)
	}
	key := derive(c.Password, a.Salt)
	if subtle.ConstantTimeCompare(key, a.Hash) != 1 || !exists {
		return Login{}, ErrCredentials
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := make(map[string]session, len(s.data.Sessions))
	for key, value := range s.data.Sessions {
		previous[key] = value
	}
	login, hash := s.issue(a.User)
	if err := s.save(); err != nil {
		delete(s.data.Sessions, hash)
		s.data.Sessions = previous
		return Login{}, err
	}
	return login, nil
}
func (s *Store) issue(user User) (Login, string) {
	now := s.now()
	for key, value := range s.data.Sessions {
		if !now.Before(value.ExpiresAt) {
			delete(s.data.Sessions, key)
		}
	}
	// One active session per account; a new login revokes the previous token.
	for key, value := range s.data.Sessions {
		if value.UserID == user.ID {
			delete(s.data.Sessions, key)
		}
	}
	token := randomToken()
	key := digest(token)
	expires := now.Add(24 * time.Hour)
	s.data.Sessions[key] = session{UserID: user.ID, ExpiresAt: expires}
	return Login{User: user, AccessToken: token, TokenType: "Bearer", ExpiresAt: expires}, key
}
func (s *Store) Authenticate(token string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(token) != 43 {
		return User{}, ErrCredentials
	}
	value, ok := s.data.Sessions[digest(token)]
	if !ok || !s.now().Before(value.ExpiresAt) {
		return User{}, ErrCredentials
	}
	for _, a := range s.data.Accounts {
		if a.User.ID == value.UserID {
			return a.User, nil
		}
	}
	return User{}, ErrCredentials
}
func (s *Store) Logout(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := digest(token)
	old, exists := s.data.Sessions[key]
	delete(s.data.Sessions, key)
	if err := s.save(); err != nil {
		if exists {
			s.data.Sessions[key] = old
		}
		return err
	}
	return nil
}
func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	raw, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".auth-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err = tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, s.path)
}
