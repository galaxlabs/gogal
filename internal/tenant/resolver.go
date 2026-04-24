package tenant

func ResolveSiteName(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return "example.local"
}
