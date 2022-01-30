package main

type scoredWords map[string]float64

func uniqueLetters(w string) int {
	var counts [256]bool
	for _, b := range w {
		counts[b] = true
	}

	var count int
	for _, b := range counts {
		if b {
			count++
		}
	}
	return count
}

func scoreWords2(words []string) scoredWords {
	//sw := make(scoredWords)
	return nil
}

func scoreWords(words []string) scoredWords {
	// scorer functions by summing how many times each letter occurs at a particular position

	// Generate counts
	var counts [5][256]float64
	for _, w := range words {
		for i, b := range w {
			counts[i][b]++
		}
	}

	sw := make(scoredWords)
	for _, w := range words {
		var score float64
		// Dock repeating letters

		// Disqualify as many as possible based on position
		for i, b := range w {
			// Optimize for 50% elimination

			// This one optimizes for the rarest letters
			//score += float64(len(words)) / counts[i][b]

			score += counts[i][b] / float64(len(words))
		}

		score = 1 / score

		score = float64(uniqueLetters(w)) * score

		//score = 10

		// Score for existence in word, not position
		//for _, b := range w {
		//	for _, count := range counts {
		//		score += float64(len(words)) / (float64(count[b]) * 5)
		//	}
		//}
		sw[w] = score
	}
	return sw
}
