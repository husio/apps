package unote

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

// JSONResp write content as JSON encoded response.
func JSONResp(w http.ResponseWriter, content interface{}, code int) {
	b, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		log.Printf("cannot JSON serialize response: %s", err)
		code = http.StatusInternalServerError
		b = []byte(`{"errors":["Internal Server Errror"]}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)

	const MB = 1 << (10 * 2)
	if len(b) > MB {
		log.Printf("response JSON body is huge: %d", len(b))
	}
	_, _ = w.Write(b)
}

// JSONErr write single error as JSON encoded response.
func JSONErr(w http.ResponseWriter, errText string, code int) {
	JSONErrs(w, []string{errText}, code)
}

// JSONErrs write multiple errors as JSON encoded response.
func JSONErrs(w http.ResponseWriter, errs []string, code int) {
	resp := struct {
		Code   int
		Errors []string `json:"errors"`
	}{
		Code:   code,
		Errors: errs,
	}
	JSONResp(w, resp, code)
}

// StdJSONResp write JSON encoded, standard HTTP response text for given status
// code. Depending on status, either error or successful response format is
// used.
func StdJSONResp(w http.ResponseWriter, code int) {
	if code >= 400 {
		JSONErr(w, http.StatusText(code), code)
	} else {
		JSONResp(w, http.StatusText(code), code)
	}
}

// JSONRedirect return redirect response, but with JSON formatted body.
func JSONRedirect(w http.ResponseWriter, urlStr string, code int) {
	w.Header().Set("Location", urlStr)
	var content = struct {
		Code     int
		Location string
	}{
		Code:     code,
		Location: urlStr,
	}
	JSONResp(w, content, code)
}

func generateId() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	s := base64.URLEncoding.EncodeToString(b)
	return s[:32]
}
