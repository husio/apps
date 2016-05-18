package main

import (
	"fmt"
	"net/http"

	"github.com/husio/apps/auth/keys"
	"github.com/husio/x/web"

	"golang.org/x/net/context"
)

// handlePublicKey return public RSA key for requested key ID.
func handlePublicKey(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	kid := web.Args(ctx).ByIndex(0)
	key, ok := keys.Manager(ctx).KeyByID(kid)
	if !ok {
		web.JSONErr(w, "key not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/pkix-crl")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pub"`, kid))
	fmt.Fprint(w, key)
}

func handleAuthenticate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// TODO use oauth?
}
