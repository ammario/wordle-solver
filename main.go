package main

import (
	"flag"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"go.coder.com/flog"
)

type corpus struct {
	words  []string
	index  *suffixarray.Index
	scores scoredWords
}

func main() {
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
		index:  index,
		words:  words,
		scores: scoreWords(words),
	}

	if showTop {
		type scoredWord struct {
			word  string
			score float64
		}
		scoreArr := make([]scoredWord, 0, len(c.scores))

		for word, score := range c.scores {
			scoreArr = append(scoreArr, scoredWord{
				word:  word,
				score: score,
			})
		}
		sort.Slice(scoreArr, func(i, j int) bool {
			return scoreArr[i].score < scoreArr[j].score
		})
		fmt.Printf("top words: ")
		for i := 0; i < 10; i++ {
			fmt.Printf("%v ", scoreArr[i].word)
		}

		fmt.Printf("\n")

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

	for i := 0; i < count; i++ {
		for _, w := range c.words {
			words <- w
		}
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
