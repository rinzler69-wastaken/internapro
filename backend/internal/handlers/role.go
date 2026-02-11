package handlers

// normalizeRole maps legacy "supervisor" to "pembimbing"
func normalizeRole(role string) string {
	if role == "supervisor" {
		return "pembimbing"
	}
	return role
}
