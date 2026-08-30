package inflection

import (
	"github.com/jinzhu/inflection"
)

func init() {
	inflection.AddSingular(
		`(nexus|campus|bonus|focus|virus|census|corpus|sinus|circus|`+
			`plexus|canvas|lens|atlas)(es)?$`,
		`$1`,
	)
	inflection.AddPlural(
		`(nexus|campus|bonus|focus|virus|census|corpus|sinus|circus|`+
			`plexus|canvas|lens|atlas)$`,
		`${1}es`,
	)
	inflection.AddUncountable("species")
}

// Singular converts a word to its singular form.
func Singular(word string) string {
	return inflection.Singular(word)
}

// Plural converts a word to its plural form.
func Plural(word string) string {
	return inflection.Plural(word)
}
