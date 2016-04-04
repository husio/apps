package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/husio/x/stamp"
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

func (km *KeyManager) Vault() *stamp.Vault {
	return &km.vault
}

func (km *KeyManager) GenerateKey(expireIn time.Duration) error {
	priv, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return err
	}
	kid := randStr(6)

	s := stamp.NewRSA256Signer(priv)
	km.vault.Add(kid, s, expireIn)

	oldkeys := km.keys()
	newkeys := make([]PubKey, 0, len(oldkeys))
	now := time.Now()
	repr, err := pubKeyStr(&priv.PublicKey)
	if err != nil {
		return fmt.Errorf("cannot format public key: %s", err)
	}
	newkeys = append(newkeys, PubKey{
		id:        kid,
		repr:      repr,
		validTill: now.Add(expireIn),
	})

	for _, key := range oldkeys {
		if key.validTill.After(now) && key.id != kid {
			newkeys = append(newkeys, key)
		}
	}

	log.Printf("new key generated: %s", kid)
	km.pubKeys.Store(newkeys)
	return nil
}

func (km *KeyManager) keys() []PubKey {
	pk := km.pubKeys.Load()
	if pk == nil {
		return nil
	}
	return pk.([]PubKey)
}

func (km *KeyManager) KeyByID(id string) (string, bool) {
	now := time.Now()
	for _, k := range km.keys() {
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
