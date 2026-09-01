package haiku

import (
	"fmt"
	"math/rand/v2"
)

var adjectives = []string{
	"silent", "crimson", "swift", "hidden", "frosty", "golden", "sleepy", "brave", "lively", "misty",
	"calm", "bright", "dark", "bold", "gentle", "wild", "eager", "loyal", "cool", "warm",
	"proud", "noble", "sharp", "soft", "keen", "vibrant", "ancient", "cosmic", "solar", "lunar",
	"sturdy", "clever", "fierce", "jolly", "merry", "smooth", "rough", "agile", "steady",
	"sparkling", "glowing", "radiant", "serene", "placid", "quiet", "dusk", "dawn", "spring", "autumn",
	"winter", "summer", "windy", "breezy", "stormy", "sunny", "cloudy", "muddy", "rocky", "sandy",
	"grassy", "flowery", "verdant", "azure", "emerald", "ruby", "amber", "silver", "bronze", "copper",
	"daring", "humble", "grand", "royal", "fancy", "simple", "plain", "heavy", "light", "fast", "slow",
	"tall", "short", "deep", "shallow", "wide", "narrow", "cheerful", "happy", "joyful", "lucky",
	"friendly", "kind", "honest", "wise", "smart", "graceful", "magic", "mystic", "ghostly", "hollow",
}

var nouns = []string{
	"forest", "badger", "river", "falcon", "meadow", "coyote", "summit", "shadow", "otter", "glacier",
	"canyon", "desert", "ocean", "breeze", "beacon", "castle", "temple", "crag", "spire", "vale",
	"brook", "pond", "lake", "crest", "ridge", "cliff", "dune", "grove", "woods", "jungle",
	"puma", "panther", "eagle", "hawk", "raven", "heron", "wolf", "fox", "bear", "deer",
	"comet", "meteor", "planet", "nebula", "galaxy", "orbit", "pulse", "wave", "spark", "flame",
	"mountain", "valley", "hill", "peak", "island", "cave", "stone", "rock", "pebble", "boulder",
	"stream", "creek", "waterfall", "spring", "geyser", "swamp", "marsh", "bog", "fen", "moor",
	"woodland", "orchard", "garden", "field", "plain", "prairie", "savanna", "tundra", "steppe", "oasis",
	"lion", "tiger", "leopard", "jaguar", "cheetah", "lynx", "bobcat", "cougar", "owl", "sparrow",
	"finch", "osprey", "kestrel", "merlin", "harrier", "kite", "buzzard", "goshawk", "sparrowhawk", "condor",
}

// Generate creates a random Heroku-style identifier with a 4-char hex suffix.
// Example: "frosty-otter-2a3f"
func Generate() string {
	adj := adjectives[rand.N(len(adjectives))]
	noun := nouns[rand.N(len(nouns))]
	suffix := rand.Uint32() & 0xffff // 4 hex characters
	return fmt.Sprintf("%s-%s-%04x", adj, noun, suffix)
}
