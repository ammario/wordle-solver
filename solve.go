package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coder/flog"
)

type hint int

const (
	hintNone   = 0
	hintInWord = 1
	hintRight  = 2
)

type letterHint struct {
	letter byte
	hint   hint
}

type lineHint [5]letterHint

func (h lineHint) Word() string {
	var s strings.Builder
	for _, c := range h {
		s.WriteByte(c.letter)
	}
	return s.String()
}

func (h lineHint) String() string {
	var s strings.Builder
	for k := range h {
		switch h[k].hint {
		case hintNone:
			s.WriteString("⬜️")
		case hintInWord:
			s.WriteRune('🟨')
		case hintRight:
			s.WriteRune('🟩')
		}
	}
	return s.String()
}

func (h lineHint) won() bool {
	for _, v := range h {
		if v.hint != hintRight {
			return false
		}
	}
	return true
}

type puzzle struct {
	hardMode bool
	hints    [][5]letterHint
}

func (p puzzle) turn() int {
	return len(p.hints)
}

func (p *puzzle) copy() *puzzle {
	pp := &puzzle{}
	for _, h := range p.hints {
		pp.hints = append(pp.hints, h)
	}
	return pp
}

func fastIndex(str string, search byte) int {
	switch {
	case str[0] == search:
		return 0
	case str[1] == search:
		return 1
	case str[2] == search:
		return 2
	case str[3] == search:
		return 3
	case str[4] == search:
		return 4
	default:
		return -1
	}
	panic("oops")
}

// Guess does not mutate possibilities.
func (p *puzzle) guess(recur bool, c *corpus, ps map[string]struct{}, deletes []string) (string, []string) {
	if p.turn() == 0 {
		return c.firstGuess, nil
	}
	// Excise invalid guesses.
	for i := 0; i < p.turn(); i++ {
		lineHint := p.hints[i]
	wordLoop:
		for word := range ps {
			// Artifact
			if len(word) != 5 {
				deletes = append(deletes, word)
				continue
			}
			for i, letterHint := range lineHint {
				switch letterHint.hint {
				case hintNone:
					if fastIndex(word, letterHint.letter) >= 0 {
						deletes = append(deletes, word)
						continue wordLoop
					}
				case hintInWord:
					if i := fastIndex(word, letterHint.letter); i < 0 {
						deletes = append(deletes, word)
						continue wordLoop
					}
					if word[i] == letterHint.letter {
						deletes = append(deletes, word)
						continue wordLoop
					}
				case hintRight:
					if word[i] != letterHint.letter {
						deletes = append(deletes, word)
						continue wordLoop
					}
				}
			}
		}
	}

	if !recur {
		for guess := range ps {
			return guess, deletes
		}
	} else {
		for _, d := range deletes {
			//fmt.Printf("delete: %v\n", d)
			delete(ps, d)
		}
	}

	if len(ps) == 1 {
		for guess := range ps {
			return guess, deletes
		}
	}

	type scoredGuess struct {
		score float64
		guess string
	}

	var (
		scoredGuessesMu sync.Mutex
		scoredGuesses   []scoredGuess
		wg              = sync.WaitGroup{}
		guesses         = make(chan string)
	)

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			deletes := make([]string, 0, 1024)

			for guess := range guesses {
				var score float64
				//start := time.Now()
				for maybeSecret := range ps {
					pp := p.copy()
					pp.hints = append(pp.hints, giveHint(maybeSecret, guess))
					_, deletes = pp.guess(false, c, ps, deletes[:0])
					score += float64(len(deletes))
					//fmt.Printf("%v > %v | %v %v\n", guess, maybeSecret, len(deletes), len(ps))
				}
				//fmt.Printf("%v | d=%v | %v | %v\n", guess, score, time.Since(start), len(ps))
				scoredGuessesMu.Lock()
				scoredGuesses = append(scoredGuesses, scoredGuess{
					score: score,
					guess: guess,
				})
				scoredGuessesMu.Unlock()
			}
		}()
	}
	// For each word, pretend every other remaining possibility is the secret.
	// Which guess has the greatest elimination?
	if p.hardMode {
		for guess := range ps {
			guesses <- guess
		}
	} else {
		for _, guess := range c.words {
			guesses <- guess
		}
	}
	close(guesses)

	wg.Wait()

	var (
		bestGuess string
		bestScore float64
	)
	for _, gs := range scoredGuesses {
		if gs.score >= bestScore {
			bestScore, bestGuess = gs.score, gs.guess
		}
	}
	if bestGuess == "" {
		panic("no guess?")
	}
	if len(ps) == 1 {
		fmt.Println(bestGuess, bestScore)
		fmt.Println(ps)
	}

	return bestGuess, deletes
}

func giveHint(secret string, guess string) lineHint {
	var hint lineHint
	for i := 0; i < 5; i++ {
		hint[i].letter = guess[i]
		switch {
		case secret[i] == guess[i]:
			hint[i].hint = hintRight
		case strings.Contains(secret, string(guess[i])):
			hint[i].hint = hintInWord
		default:
			hint[i].hint = hintNone
		}
	}
	return hint
}

func solve(log *flog.Logger, secret string, hardMode bool, c *corpus) int {
	p := puzzle{hardMode: hardMode}

	// Copy the possibilities array for modification.
	possibilities := make(map[string]struct{}, len(c.words))
	for _, w := range c.words {
		possibilities[w] = struct{}{}
	}

	var (
		guess   string
		deletes []string
	)

	deletes = make([]string, 0, len(possibilities))

	for {
		start := time.Now()
		guess, deletes = p.guess(true, c, possibilities, deletes)
		for _, d := range deletes {
			delete(possibilities, d)
		}
		delete(possibilities, guess)
		h := giveHint(secret, guess)
		log.Info("rem %04d -> %v: %v | took %v", len(possibilities)+1, guess, h.String(), time.Since(start))
		if h.won() {
			log.Success("%q found in %v guesses :)", secret, p.turn()+1)
			return p.turn()
		}
		p.hints = append(p.hints, h)
	}
}
