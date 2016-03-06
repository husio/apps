package webcron

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"sync/atomic"
)

var (
	instanceID string = randomID()
	counter    int64  = 0
)

func randomID() string {
	const length = 8

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic("cannot read random value: " + err.Error())
	}
	s := base32.StdEncoding.EncodeToString(b)[:length]
	return strings.ToLower(s)
}

// genID return unique ID string
func genID() string {
	n := atomic.AddInt64(&counter, 1)
	return fmt.Sprintf("%s:%d", instanceID, n)
}
