package ids

import gonanoid "github.com/matoous/go-nanoid/v2"

// SlugLength is the length of capability slugs. 21 URL-safe chars ≈ 126 bits
// of entropy — knowing the slug is what grants access to a list.
const SlugLength = 21

// NewSlug generates a URL-safe capability slug.
func NewSlug() (string, error) {
	return gonanoid.New(SlugLength)
}
