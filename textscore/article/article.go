package article

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

const (
	minWordCount = 10
)

var ignoretags = map[string]struct{}{
	"script": struct{}{},
	"style":  struct{}{},
	"head":   struct{}{},
	"footer": struct{}{},
	"form":   struct{}{},
}

var nooptags = map[string]struct{}{
	"h1":     struct{}{},
	"h2":     struct{}{},
	"h3":     struct{}{},
	"h4":     struct{}{},
	"h5":     struct{}{},
	"h6":     struct{}{},
	"em":     struct{}{},
	"strong": struct{}{},
	"span":   struct{}{},

	"a": struct{}{},
}

func isnooptag(tag string) bool {
	_, ok := nooptags[tag]
	return ok
}

func Parse(r io.Reader, w io.Writer) error {
	cns, err := extractCandidates(r)
	if err != nil {
		return err
	}

	if len(cns) == 0 {
		return ErrNoArticle
	}

	sort.Sort(cns)

	cn := cns[0]
	for _, s := range cn.chunks {
		if _, err := fmt.Fprintln(w, s); err != nil {
			return err
		}
	}
	fmt.Println(cn.score)
	return nil
}

var ErrNoArticle = errors.New("no article")

func extractCandidates(r io.Reader) (candidates, error) {
	tokenizer := html.NewTokenizer(r)

	var candidates candidates
	var c *candidate

	var ignore int
	var nodepos int
	var stack []*html.Token

parseDocument:
	for {
		switch tp := tokenizer.Next(); tp {
		case html.ErrorToken:
			break parseDocument

		case html.TextToken:
			if ignore > 0 {
				continue
			}
			token := tokenizer.Token()

			var tag string
			if len(stack) > 0 {
				tag = stack[len(stack)-1].Data
			}
			wc := wordcount(token.Data)
			switch {
			case c != nil && isnooptag(tag):
				// all good
			case wc < minWordCount:
				continue parseDocument
			}

			if c == nil {
				c = &candidate{
					score:   -(len(candidates) * 2),
					nodepos: nodepos,
					chunks:  make([]string, 0, 8),
				}
				candidates = append(candidates, c)
			}

			if tag == "a" {
				c.lwc += wordcount(token.Data)
			} else {
				c.score += (wc / 50)
				c.wc += wc
				c.score += strings.Count(token.Data, ",")
			}

			c.chunks = append(c.chunks, token.Data)

		case html.StartTagToken:
			token := tokenizer.Token()

			nodepos++
			stack = append(stack, &token)

			if _, ok := ignoretags[token.Data]; ok {
				ignore++
				continue
			}

		case html.EndTagToken:
			stack = stack[:len(stack)-1]

			token := tokenizer.Token()

			if _, ok := ignoretags[token.Data]; ok {
				ignore--
			}

			if c != nil && nodepos > c.nodepos-20 {
				c.score = int(float64(c.score) * (1.1 - (float64(c.lwc) / float64(c.wc))))
				c = nil
			}

		case html.SelfClosingTagToken:
		case html.CommentToken:
		case html.DoctypeToken:
		}
	}
	return candidates, nil
}

type candidate struct {
	nodepos int
	chunks  []string
	wc      int
	lwc     int
	score   int
}

type candidates []*candidate

func (cns candidates) Len() int           { return len(cns) }
func (cns candidates) Less(i, j int) bool { return cns[i].score > cns[j].score }
func (cns candidates) Swap(i, j int)      { cns[i], cns[j] = cns[j], cns[i] }

func wordcount(text string) int {
	return len(strings.Fields(text))
}
