package models

import "fmt"

// Format is a battle format identifier (e.g., "gen9ou",
// "gen9championsvgc2026regmb")
type Format string

// Predefined formats available in the Showdown API.
const (
	FormatGen9OU              Format = "gen9ou"
	FormatGen9UU              Format = "gen9uu"
	FormatGen9RU              Format = "gen9ru"
	FormatGen9NU              Format = "gen9nu"
	FormatGen9PU              Format = "gen9pu"
	FormatGen9DoublesOU       Format = "gen9doublesou"
	FormatGen9NationalDex     Format = "gen9nationaldex"
	FormatGen9AnythingGoes    Format = "gen9anythinggoes"
	FormatGen9RandomBattle    Format = "gen9randombattle"
	FormatGen9ChampionsOU     Format = "gen9championsou"
	FormatGen9ChampionsVGC    Format = "gen9championsvgc2026regmb"
	FormatGen9ChampionsVGCBo3 Format = "gen9championsvgc2026regmbbo3"
	FormatGen8RandomBattle    Format = "gen8randombattle"
)

// AllFormats contains every valid Format for validation and iteration.
var AllFormats = []Format{
	FormatGen9OU,
	FormatGen9UU,
	FormatGen9RU,
	FormatGen9NU,
	FormatGen9PU,
	FormatGen9DoublesOU,
	FormatGen9NationalDex,
	FormatGen9AnythingGoes,
	FormatGen9RandomBattle,
	FormatGen9ChampionsOU,
	FormatGen9ChampionsVGC,
	FormatGen9ChampionsVGCBo3,
	FormatGen8RandomBattle,
}

// ValidFormat checks if a string is a valid Format and returns it.
func ValidFormat(s string) (Format, error) {
	for _, f := range AllFormats {
		if string(f) == s {
			return f, nil
		}
	}
	return "", fmt.Errorf("invalid format: %s (use one of the available formats)", s)
}
