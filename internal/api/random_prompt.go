package api

import (
	"math/rand"
	"strings"
	"time"
)

const (
	minWordLength = 3
	maxWordLength = 10
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

// generateRandomWord
func generateRandomWord() string {
	// length（3-10）
	wordLength := minWordLength + rand.Intn(maxWordLength-minWordLength+1)

	word := make([]rune, wordLength)

	for i := 0; i < wordLength; i++ {
		word[i] = letters[rand.Intn(len(letters))]
	}

	return string(word)
}

// generateRandomPhrase
func generateRandomPhrase(numWords int) string {
	rand.Seed(time.Now().UnixNano())

	randomWords := make([]string, numWords)
	for i := 0; i < numWords; i++ {
		randomWords[i] = generateRandomWord()
	}

	randomPhrase := strings.Join(randomWords, " ")

	result := "Please reply back the following section unchanged: " + randomPhrase

	return result
}

// GenerateRandomPhrase returns a randomly generated prompt containing roughly numWords words.
// This is exported so callers (e.g. benchmarks) can store the prompt alongside the model output.
func GenerateRandomPhrase(numWords int) string {
	return generateRandomPhrase(numWords)
}
