package bb

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/husio/x/log"
)

const baseTmplPath = "bb/templates/base.tmpl"

var baseFuncs = template.FuncMap{
	"join": strings.Join,
}

func render(w io.Writer, templateName string, context interface{}) {
	// XXX cache

	t, err := parseFiles(baseTmplPath, "bb/templates/"+templateName)
	if err != nil {
		log.Error("cannot parse template",
			"template", templateName,
			"error", err.Error())
		return
	}
	t = t.Funcs(baseFuncs)

	if err := t.Execute(w, context); err != nil {
		log.Error("cannot render template",
			"template", templateName,
			"error", err.Error())
		return
	}
}

func respond500(w http.ResponseWriter, r *http.Request) {
	// XXX: detect content type
	w.WriteHeader(500)
	fmt.Fprint(w, "500")
}

func respond404(w http.ResponseWriter, r *http.Request) {
	// XXX: detect content type
	w.WriteHeader(404)
	fmt.Fprint(w, "404")
}

func parseFiles(filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}

	var t *template.Template
	for _, filename := range filenames {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		name := filepath.Base(filename)
		var tmpl *template.Template
		if t == nil {
			t = template.New(name).Funcs(baseFuncs)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
