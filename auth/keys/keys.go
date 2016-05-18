package keys

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/husio/x/stamp"
	"golang.org/x/net/context"
)

var rsaKeySize = 1024 * 2

type KeyManager struct {
	vault   stamp.Vault
	pubKeys atomic.Value
}

type PubKey struct {
	id        string
	repr      string
	validTill time.Time
}

func (m *KeyManager) Vault() *stamp.Vault {
	return &m.vault
}

func (m *KeyManager) Add(id string, key *rsa.PrivateKey, expireIn time.Duration) error {
	s := stamp.NewRSA256Signer(key)
	m.vault.AddSigner(id, s, expireIn)

	oldkeys := m.keys()
	newkeys := make([]PubKey, 0, len(oldkeys))
	now := time.Now()
	repr, err := pubKeyStr(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("cannot format public key: %s", err)
	}
	newkeys = append(newkeys, PubKey{
		id:        id,
		repr:      repr,
		validTill: now.Add(expireIn),
	})

	for _, key := range oldkeys {
		if key.validTill.After(now) && key.id != id {
			newkeys = append(newkeys, key)
		}
	}

	m.pubKeys.Store(newkeys)
	return nil
}

func (m *KeyManager) GenerateKey(expireIn time.Duration) (string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return "", err
	}
	kid := randStr(8)
	if err := m.Add(kid, priv, expireIn); err != nil {
		return "", err
	}
	return kid, nil
}

func (m *KeyManager) keys() []PubKey {
	pk := m.pubKeys.Load()
	if pk == nil {
		return nil
	}
	return pk.([]PubKey)
}

func (m *KeyManager) KeyByID(id string) (string, bool) {
	now := time.Now()
	for _, k := range m.keys() {
		if k.id != id {
			continue
		}
		if k.validTill.Before(now) {
			return "", false
		}
		return k.repr, true
	}
	return "", false
}

func randStr(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:size]
}

func pubKeyStr(key *rsa.PublicKey) (string, error) {
	raw, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", fmt.Errorf("cannot marshal key: %s", err)
	}

	block := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   raw,
	}

	var b bytes.Buffer
	if err := pem.Encode(&b, &block); err != nil {
		return "", fmt.Errorf("cannot encode block: %s", err)
	}
	return b.String(), nil
}

func WithManager(ctx context.Context, m *KeyManager) context.Context {
	return context.WithValue(ctx, "auth:keymanager", m)
}

func Manager(ctx context.Context) *KeyManager {
	m := ctx.Value("auth:keymanager")
	if m == nil {
		panic("key manager not present in context")
	}
	return m.(*KeyManager)
}
