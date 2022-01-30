package main

import (
	"fmt"
	"io"
	"strings"
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
	turn  int
	hints [][5]letterHint
}

func (p *puzzle) guess(c *corpus, possibilities map[string]struct{}) string {
	// Excise invalid guesses.
	for i := 0; i < p.turn; i++ {
		lineHint := p.hints[i]
		for i, letterHint := range lineHint {
			for word := range possibilities {
				// Artifact
				if len(word) != 5 {
					delete(possibilities, word)
					continue
				}
				switch letterHint.hint {
				case hintNone:
					if strings.Contains(word, string(letterHint.letter)) {
						delete(possibilities, word)
					}
				case hintInWord:
					if !strings.Contains(word, string(letterHint.letter)) {
						delete(possibilities, word)
					}
				case hintRight:
					if word[i] != letterHint.letter {
						delete(possibilities, word)
					}
				}
			}
		}
	}

	var (
		bestGuess string
		bestScore float64
	)
	for word := range possibilities {
		score := c.scores[word]
		if score >= bestScore {
			bestGuess, bestScore = word, score
		}
	}

	delete(possibilities, bestGuess)
	return bestGuess
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

func solve(log io.Writer, secret string, c *corpus) int {
	p := puzzle{}

	// Copy the possibilities array for modification.
	possibilities := make(map[string]struct{}, len(c.words))
	for _, w := range c.words {
		possibilities[w] = struct{}{}
	}

	for {
		guess := p.guess(c, possibilities)
		h := giveHint(secret, guess)
		fmt.Fprintf(log, "%v: %v\n", guess, h.String())
		if h.won() {
			fmt.Fprintf(log, "%q found in %v guesses :)\n", secret, p.turn+1)
			return p.turn
		}
		p.hints = append(p.hints, h)
		p.turn++
	}
}
