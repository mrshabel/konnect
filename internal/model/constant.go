package model

var (
	// system-default interests
	SystemInterests = []string{
		// arts & culture
		"Art", "Photography", "Fashion", "Content Creation", "TikTok", "Anime", "Creative Direction",

		// entertainment & media
		"Netflix", "Movies", "Reality TV", "YouTube", "Comedy Skits", "Podcasts",

		// food & drink
		"Foodie", "Brunch", "Cooking", "Street Food", "Cocktails", "Jollof", "Waakye",

		// music & audio
		"Afrobeats", "Amapiano", "Highlife", "Hip Hop", "R&B", "Live Music", "DJ Nights",

		// outdoors & travel
		"Travel", "Beach Hangouts", "Road Trips", "Aworshia", "Volta Trips",

		// sports & fitness
		"Gym", "Jogging", "Football", "Dance Workouts", "Yoga",

		// lifestyle & self-care
		"Self Care", "Skincare", "Meditation", "Fashion Forward", "Thrifting",

		// social & events
		"Detty December", "Tidal Rave", "AfroFuture", "Nightlife", "House Parties", "Sip & Paint", "Game Nights",

		// career & modern interests
		"Tech & Coding", "Entrepreneurship", "Startups", "Finance",
	}

	// validInterestsSet is a read-only lookup for validation
	validInterestsSet = make(map[string]struct{}, len(SystemInterests))
)

// seed the lookup map once at startup
func init() {
	for _, interest := range SystemInterests {
		validInterestsSet[interest] = struct{}{}
	}
}

func IsValidInterest(interest string) bool {
	_, ok := validInterestsSet[interest]
	return ok
}

func ValidateInterests(interests []string) bool {
	for _, interest := range interests {
		if !IsValidInterest(interest) {
			return false
		}
	}
	return true
}
