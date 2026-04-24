package crud

import "strings"

func SanitizeIdentifier(id string) string {
	id = strings.TrimSpace(strings.ToLower(id))
	id = strings.ReplaceAll(id, "-", "_")
	id = strings.ReplaceAll(id, " ", "_")
	builder := strings.Builder{}
	for _, ch := range id {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			builder.WriteRune(ch)
		}
	}
	return strings.Trim(builder.String(), "_")
}
