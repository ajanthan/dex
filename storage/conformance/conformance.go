// +build go1.7

// Package conformance provides conformance tests for storage implementations.
package conformance

import (
	"reflect"
	"testing"
	"time"

	"github.com/coreos/dex/storage"

	"github.com/kylelemons/godebug/pretty"
)

var neverExpire = time.Now().Add(time.Hour * 24 * 365 * 100)

// StorageFactory is a method for creating a new storage. The returned storage sould be initialized
// but shouldn't have any existing data in it.
type StorageFactory func() storage.Storage

// RunTestSuite runs a set of conformance tests against a storage.
func RunTestSuite(t *testing.T, sf StorageFactory) {
	tests := []struct {
		name string
		run  func(t *testing.T, s storage.Storage)
	}{
		{"UpdateAuthRequest", testUpdateAuthRequest},
		{"CreateRefresh", testCreateRefresh},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.run(t, sf())
		})
	}
}

func testUpdateAuthRequest(t *testing.T, s storage.Storage) {
	a := storage.AuthRequest{
		ID:            storage.NewID(),
		ClientID:      "foobar",
		ResponseTypes: []string{"code"},
		Scopes:        []string{"openid", "email"},
		RedirectURI:   "https://localhost:80/callback",
		Expiry:        neverExpire,
	}

	identity := storage.Claims{Email: "foobar"}

	if err := s.CreateAuthRequest(a); err != nil {
		t.Fatalf("failed creating auth request: %v", err)
	}
	if err := s.UpdateAuthRequest(a.ID, func(old storage.AuthRequest) (storage.AuthRequest, error) {
		old.Claims = identity
		old.ConnectorID = "connID"
		return old, nil
	}); err != nil {
		t.Fatalf("failed to update auth request: %v", err)
	}

	got, err := s.GetAuthRequest(a.ID)
	if err != nil {
		t.Fatalf("failed to get auth req: %v", err)
	}
	if !reflect.DeepEqual(got.Claims, identity) {
		t.Fatalf("update failed, wanted identity=%#v got %#v", identity, got.Claims)
	}
}

func testCreateRefresh(t *testing.T, s storage.Storage) {
	id := storage.NewID()
	refresh := storage.RefreshToken{
		RefreshToken: id,
		ClientID:     "client_id",
		ConnectorID:  "client_secret",
		Scopes:       []string{"openid", "email", "profile"},
	}
	if err := s.CreateRefresh(refresh); err != nil {
		t.Fatalf("create refresh token: %v", err)
	}
	gotRefresh, err := s.GetRefresh(id)
	if err != nil {
		t.Fatalf("get refresh: %v", err)
	}
	if !reflect.DeepEqual(gotRefresh, refresh) {
		t.Errorf("refresh returned did not match expected")
	}

	if err := s.DeleteRefresh(id); err != nil {
		t.Fatalf("failed to delete refresh request: %v", err)
	}

	if _, err := s.GetRefresh(id); err != storage.ErrNotFound {
		t.Errorf("after deleting refresh expected storage.ErrNotFound, got %v", err)
	}

}

func mustBeErrNotFound(t *testing.T, kind string, err error) {
	switch {
	case err == nil:
		t.Errorf("deleting non-existant %s should return an error", kind)
	case err != storage.ErrNotFound:
		t.Errorf("deleting %s expected storage.ErrNotFound, got %v", kind, err)
	}
}

func testCreateClient(t *testing.T, s storage.Storage) {
	id := storage.NewID()
	c := storage.Client{
		ID:           id,
		Secret:       "foobar",
		RedirectURIs: []string{"foo://bar.com/", "https://auth.example.com"},
		Name:         "dex client",
		LogoURL:      "https://goo.gl/JIyzIC",
	}
	err := s.DeleteClient(id)
	mustBeErrNotFound(t, "client", err)

	if err := s.CreateClient(c); err != nil {
		t.Fatalf("create client: %v", err)
	}

	getAndCompare := func(id string, want storage.Client) {
		gc, err := s.GetClient(id)
		if err != nil {
			t.Errorf("get client: %v", err)
			return
		}
		if diff := pretty.Compare(c, gc); diff != "" {
			t.Errorf("client retrieved from storage did not match: %s", diff)
		}
	}

	getAndCompare(id, c)

	newSecret := "barfoo"
	err = s.UpdateClient(id, func(old storage.Client) (storage.Client, error) {
		old.Secret = newSecret
		return old, nil
	})
	if err != nil {
		t.Errorf("update client: %v", err)
	}
	c.Secret = newSecret
	getAndCompare(id, c)

	if err := s.DeleteClient(id); err != nil {
		t.Fatalf("delete client: %v", err)
	}

	_, err = s.GetClient(id)
	mustBeErrNotFound(t, "client", err)
}
