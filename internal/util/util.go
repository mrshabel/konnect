package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var (
	// adjectives for matchmaking - positive, flirty, fun
	adjectives = []string{
		"Charming", "Witty", "Playful", "Sweet", "Bold",
		"Daring", "Clever", "Warm", "Bright", "Lucky",
		"Cosmic", "Golden", "Velvet", "Silver", "Smooth",
		"Cool", "Wild", "Gentle", "Rare", "Pure",
		"Mystic", "Epic", "Noble", "Royal", "Divine",
		"Secret", "Dreamy", "Sunny", "Starry", "Moonlit",
	}

	// nouns for matchmaking - animals, nature, cosmic
	nouns = []string{
		"Sankofa", "Adinkra", "Kente", "Akoma", "Nyame",
		"Dwennimmen", "Eban", "Mmere", "Fihankra", "Gye",
		"Lion", "Elephant", "Leopard", "Eagle", "Falcon",
		"Antelope", "Cheetah", "Gazelle", "Hawk", "Phoenix",
		"Sahara", "Savanna", "Baobab", "Acacia", "Harmattan",
		"Star", "Moon", "Sun", "Thunder", "Storm",
		"Gold", "Diamond", "Amber", "Ivory", "Pearl",
		"Coral", "Ruby", "Jade", "Opal", "Crystal",
	}
)

// GenerateRandomUsername creates a random nickname in the format "AdjectiveNoun42"
func GenerateRandomUsername() string {
	adjective := randomElement(adjectives)
	noun := randomElement(nouns)

	// random number between 10 and 99
	num := randomNumber(10, 999)

	return fmt.Sprintf("%s%s%d", adjective, noun, num)
}

// randomElement returns a random element from a slice
func randomElement(slice []string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(slice))))
	return slice[n.Int64()]
}

// randomNumber generates a random number between min and max
func randomNumber(min, max int) int {
	// generate random number
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}
