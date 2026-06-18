package auth

import "strings"

// Scope is a permission string in the format "<resource>:<action>" or
// "<resource>:<id>:<action>" for fine-grained control.
type Scope string

// ScopeMatch checks whether a required scope is covered by any of the
// granted scopes. Supports wildcard (`*`) at any segment position.
func ScopeMatch(required Scope, granted []Scope) bool {
	reqParts := splitScope(string(required))

	for _, g := range granted {
		if string(g) == "*" {
			return true
		}
		grantParts := splitScope(string(g))
		if matchSegments(reqParts, grantParts) {
			return true
		}
	}
	return false
}

func splitScope(s string) []string {
	return strings.Split(s, ":")
}

func matchSegments(required, granted []string) bool {
	for i := 0; i < len(granted); i++ {
		if granted[i] == "*" {
			return true
		}
		if i >= len(required) {
			return false
		}
		if granted[i] != required[i] {
			return false
		}
	}
	return len(required) == len(granted)
}