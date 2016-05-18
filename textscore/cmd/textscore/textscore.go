package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/husio/x/log"
)

func main() {
	minRepFl := flag.Int("minrep", 2, "Minimum repetition amount for word to be relevant")
	minWLenFl := flag.Int("minwlen", 3, "Minimum word length")
	stopwFl := flag.String("stopw", "", "Stopwords list")
	flag.Parse()

	stopw := make(map[string]struct{})
	if *stopwFl != "" {
		stopw = stopwords(*stopwFl)
	}

	counts := make(map[string]int)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		w := strings.ToLower(scanner.Text())
		if strings.HasPrefix(w, "<") || strings.HasSuffix(w, ">") {
			continue
		}
		w = strings.TrimRight(w, ",.")

		if len(w) > 40 {
			continue
		}

		if len(w) < *minWLenFl {
			continue
		}

		if _, ok := stopw[w]; ok {
			continue
		}

		counts[w]++
	}

	if err := scanner.Err(); err != nil {
		log.Error("scanner error", "error", err.Error())
	}

	var pairs pairs
	for word, count := range counts {
		if count >= *minRepFl {
			pairs = append(pairs, pair{word, count})
		}
	}

	sort.Sort(pairs)

	for _, pair := range pairs {
		fmt.Printf("%s\t%d\n", pair.word, pair.count)
	}
}

func stopwords(path string) map[string]struct{} {
	stopw := make(map[string]struct{})
	fd, err := os.Open(path)
	if err != nil {
		log.Error("cannot open stopwords file", "error", err.Error())
		return stopw
	}
	defer fd.Close()

	rd := bufio.NewReader(fd)
	for {
		word, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Error("cannot read stopwords", "error", err.Error())
			}
			return stopw
		}
		stopw[strings.TrimSpace(word)] = struct{}{}
	}
}

type pair struct {
	word  string
	count int
}

type pairs []pair

func (p pairs) Len() int           { return len(p) }
func (p pairs) Less(i, j int) bool { return p[i].count > p[j].count }
func (p pairs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
