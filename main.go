package main

import (
	"flag"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"go.coder.com/flog"
)

type corpus struct {
	words []string
	index *suffixarray.Index
}

func main() {
	wordsFile, err := ioutil.ReadFile("words.txt")
	if err != nil {
		flog.Fatal("open: %v", err)
	}

	var testFlag bool

	flag.BoolVar(&testFlag, "test", false, "test algorithm performance")

	flag.Parse()

	index := suffixarray.New(wordsFile)

	words := make([]string, 0, 6000)
	// Clean
	{
		dirtyWords := strings.Split(string(wordsFile), "\n")
		for _, w := range dirtyWords {
			if len(w) != 5 {
				continue
			}
			words = append(words, w)
		}
	}
	c := &corpus{
		index: index,
		words: words,
	}

	if !testFlag {
		secret := flag.Arg(0)
		if len(secret) != 5 {
			flog.Fatal("%q not 5 letters", secret)
		}
		if len(index.Lookup([]byte(secret), 1)) == 0 {
			flog.Fatal("%q not found in list", secret)
		}

		flog.Info("solving for %q...", secret)
		solve(os.Stdout, secret, c)
	}

	test(c)
}

func test(c *corpus) {

	var (
		turnCount uint64
		wg        sync.WaitGroup
		words     = make(chan string)

		histogramMu sync.Mutex
		histogram   [18][]string
	)

	for i := 0; i < 8; i++ {
		go func() {
			for w := range words {
				turns := solve(ioutil.Discard, w, c)
				atomic.AddUint64(&turnCount, uint64(turns))

				histogramMu.Lock()
				histogram[turns] = append(histogram[turns], w)
				histogramMu.Unlock()
			}
		}()
	}

	for _, w := range c.words {
		words <- w
	}
	close(words)
	wg.Wait()

	flog.Info("average solution in %04.2f steps", float64(turnCount)/float64(len(c.words)))
	for i, h := range histogram {
		fmt.Printf("in turns %2d: %05.2f%% | ", i+1, float64(len(h))/float64(len(c.words))*100)
		for i := 0; i < len(h) && i < 7; i++ {
			fmt.Printf("%v ", h[i])
		}
		fmt.Printf("\n")
	}
}
