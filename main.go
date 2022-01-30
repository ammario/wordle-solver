package main

import (
	"flag"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.coder.com/flog"
)

type corpus struct {
	words []string
	index *suffixarray.Index
}

func main() {
	go func() {
		log.Fatal(http.ListenAndServe("localhost:6060", nil))
	}()

	wordsFile, err := ioutil.ReadFile("words.txt")
	if err != nil {
		flog.Fatal("open: %v", err)
	}

	var (
		testFlag      int
		showHistogram bool
		showTop       bool
	)

	flag.IntVar(&testFlag, "test", 0, "test algorithm performance")
	flag.BoolVar(&showHistogram, "hist", false, "show histogram with test")
	flag.BoolVar(&showTop, "show-top", false, "show top words")

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

	if testFlag == 0 {
		secret := flag.Arg(0)
		if len(secret) != 5 {
			flog.Fatal("%q not 5 letters", secret)
		}
		if len(index.Lookup([]byte(secret), 1)) == 0 {
			flog.Fatal("%q not found in list", secret)
		}

		flog.Info("solving for %q...", secret)
		solve(os.Stdout, secret, c)
	} else {
		test(testFlag, c, showHistogram)
	}
}

func test(count int, c *corpus, showHistogram bool) {
	var (
		solves    uint64
		turnCount uint64
		wg        sync.WaitGroup
		words     = make(chan string)

		histogramMu sync.Mutex
		histogram   [25][]string
	)

	for i := 0; i < 8; i++ {
		go func() {
			for w := range words {
				turns := solve(ioutil.Discard, w, c)
				atomic.AddUint64(&turnCount, uint64(turns))
				atomic.AddUint64(&solves, 1)

				histogramMu.Lock()
				histogram[turns] = append(histogram[turns], w)
				histogramMu.Unlock()
			}
		}()
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < count; i++ {
		words <- c.words[rand.Intn(len(c.words))]
	}
	close(words)
	wg.Wait()

	fmt.Printf("average solution in %04.5f steps\n", float64(turnCount)/float64(solves))
	if !showHistogram {
		return
	}
	for i, h := range histogram {
		fmt.Printf("in turns %2d: %05.2f%% | ", i+1, float64(len(h))/float64(solves)*100)
		for i := 0; i < len(h) && i < 7; i++ {
			fmt.Printf("%v ", h[i])
		}
		fmt.Printf("\n")
	}
}
