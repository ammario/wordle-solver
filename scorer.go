package main

type scoredWords map[string]float64

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
		// Disqualify as many as possible based on position
		for i, b := range w {
			// Optimize for 50% elimination

			// This one optimizes for the rarest letters
			//score += float64(len(words)) / counts[i][b]

			score += float64(len(words)) / counts[i][b]
		}

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
