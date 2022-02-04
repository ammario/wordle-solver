package main

import (
	_ "embed"
	"flag"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/flog"
)

type corpus struct {
	firstGuess string
	words      []string
	index      *suffixarray.Index
}

//go:embed words.txt
var wordsFile []byte

func main() {
	go func() {
		_ = http.ListenAndServe("localhost:6060", nil)
	}()

	var (
		testFlag      int
		showHistogram bool
		showTop       bool
		hardMode      bool
		firstGuess    string
	)

	flag.StringVar(&firstGuess, "first-guess", "tares", "first word guess")
	flag.IntVar(&testFlag, "test", 0, "test algorithm performance")
	flag.BoolVar(&showHistogram, "hist", false, "show histogram with test")
	flag.BoolVar(&showTop, "show-top", false, "show top words")
	flag.BoolVar(&hardMode, "hard", false, "must use prior hints")

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
		firstGuess: firstGuess,
		index:      index,
		words:      words,
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
		solve(flog.New(), secret, hardMode, c)
	} else {
		test(testFlag, c, hardMode, showHistogram)
	}
}

func test(count int, c *corpus, hardMode bool, showHistogram bool) {
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
				turns := solve(&flog.Logger{W: ioutil.Discard}, w, hardMode, c)
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
