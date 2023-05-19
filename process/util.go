package process

import (
	"strings"

	"github.com/google/uuid"
)

func idstring(id uuid.UUID) string {
	s := id.String()
	return strings.ReplaceAll(s, "-", "")
}

func getUrlPath(path string) string {
	p := strings.ToLower(strings.ReplaceAll(path, " ", "-"))
	return p
}
