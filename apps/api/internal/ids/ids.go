package ids

import (
	"regexp"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// SlugLength is the length of capability slugs. 21 URL-safe chars ≈ 126 bits
// of entropy — knowing the slug is what grants access to a list.
const SlugLength = 21

// NewSlug generates a URL-safe capability slug.
func NewSlug() (string, error) {
	return gonanoid.New(SlugLength)
}

var uuidPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ValidUUID reports whether s is a well-formed UUID, so malformed identifiers
// are rejected before ever reaching the database.
func ValidUUID(s string) bool {
	return uuidPattern.MatchString(s)
}
