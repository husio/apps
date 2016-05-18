package article

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"testing"
)

var generateFl = flag.Bool("generate", false, "Generate golden data")

func TestParser(t *testing.T) {
	cases := map[string]struct {
		fixture string
		err     error
	}{
		"ok_1": {
			fixture: "fixtures/beaglebone_on_a_chip.html",
		},
		"ok_2": {
			fixture: "fixtures/when-websites-wont-take-no-for-an-answer.html",
		},
		"ok_3": {
			fixture: "fixtures/german-open-wi-fi-storehaftung-law-repealed.html",
		},
		"ok_4": {
			fixture: "fixtures/when-websites-wont-take-no-for-an-answer.html",
		},
		"ok_5": {
			fixture: "fixtures/globemail-ottowa_cuts.html",
		},
		"not_article_1": {
			fixture: "fixtures/ycombinator.html",
			err:     ErrNoArticle,
		},
	}

	for tname, tc := range cases {
		func() {
			fixfd, err := os.Open(tc.fixture)
			if err != nil {
				t.Errorf("%s: cannot open fixture: %s", tname, err)
				return
			}
			defer fixfd.Close()

			var b bytes.Buffer
			err = Parse(fixfd, &b)
			if err != tc.err {
				t.Errorf("%s: want error %#v, got %#v", tname, tc.err, err)
				return
			}
			if err != nil {
				return
			}

			respath := tc.fixture + ".gold"
			if *generateFl {
				gold, err := os.OpenFile(respath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					t.Errorf("%s: cannot create result file: %s", tname, err)
				}
				defer gold.Close()
				if _, err := b.WriteTo(gold); err != nil {
					t.Errorf("%s: cannot write golden data: %s", tname, err)
				}
				return
			}

			wantb, err := ioutil.ReadFile(respath)
			if err != nil {
				t.Errorf("%s: cannot read result file: %s", tname, err)
				return
			}
			want := string(wantb)

			if art := b.String(); art != want {
				t.Errorf(`%s:
--- want --- (below)
%q
--- got --- (below)
%q
`, tname, want, art)
			}
		}()
	}
}
