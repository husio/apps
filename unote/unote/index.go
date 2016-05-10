package unote

import (
	"html/template"
	"net/http"
	"strings"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if u := r.URL.Path; u == "/" || strings.HasPrefix(u, "/ui") {
		tmpl.Execute(w, nil)
	} else {
		StdJSONResp(w, http.StatusNotFound)
	}
}

var tmpl = template.Must(template.New("").Parse(strings.TrimSpace(`
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <meta http-equiv="X-UA-Compatible" content="IE=edge">
        <meta name="viewport" content="width=device-width, initial-scale=1">

        <link href="//cdnjs.cloudflare.com/ajax/libs/normalize/4.1.1/normalize.min.css" rel="stylesheet">
        <link href="/public/css/main.css" rel="stylesheet">

        <script src="/public/js/lib/require.js"></script>
        <script>require.config({baseUrl: "/public/js", urlArgs: "v=" + Date.now()}); require(['main'], function(main) { main() })</script>
    </head>
    <body>
        <div id="application"></div>
    </body>
</html>
`)))
