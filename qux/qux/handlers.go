package qux

import (
	"net/http"

	"golang.org/x/net/context"
)

func handleSubscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) {
}

func handleUnsubscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) {
}

func handlePong(ctx context.Context, w http.ResponseWriter, r *http.Request) {
}
