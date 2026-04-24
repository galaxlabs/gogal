package metadata

import (
	"fmt"
	"regexp"
	"strings"
)

var identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func NormalizeIdentifier(value string) string {
	v := strings.TrimSpace(strings.ToLower(value))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	b := strings.Builder{}
	for _, ch := range v {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			b.WriteRune(ch)
		}
	}
	return strings.Trim(b.String(), "_")
}

func ValidateIdentifier(value string) error {
	if !identifierPattern.MatchString(value) {
		return fmt.Errorf("identifier must start with a lowercase letter and use only lowercase letters, numbers, underscores")
	}
	return nil
}

func BuildStorageTableName(name string) string {
	id := NormalizeIdentifier(name)
	if id == "" {
		return ""
	}
	return "tab_" + id
}
