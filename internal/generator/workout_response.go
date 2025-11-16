package generator

// WorkoutResponseJSON represents the JSON structure returned by OpenAI
type WorkoutResponseJSON struct {
	Rounds []RoundResponseJSON `json:"rounds"`
}

// RoundResponseJSON represents a single round in the workout response
type RoundResponseJSON struct {
	RoundNumber int       `json:"round_number"`
	Combo       ComboJSON `json:"combo"`
}

// ComboJSON represents a combo in the JSON response
// Each combo is an array of move numbers (1-6 for punches, 7-12 for defensive moves)
type ComboJSON struct {
	Moves []int `json:"moves"`
}
