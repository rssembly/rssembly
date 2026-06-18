package auth

import "testing"

func TestScopeMatch_Exact(t *testing.T) {
	if !ScopeMatch("feeds:read", []Scope{"feeds:read"}) {
		t.Error("feeds:read should match feeds:read")
	}
}

func TestScopeMatch_WildcardAction(t *testing.T) {
	if !ScopeMatch("feeds:read", []Scope{"feeds:*"}) {
		t.Error("feeds:read should match feeds:*")
	}
	if !ScopeMatch("feeds:write", []Scope{"feeds:*"}) {
		t.Error("feeds:write should match feeds:*")
	}
}

func TestScopeMatch_Superadmin(t *testing.T) {
	if !ScopeMatch("feeds:read", []Scope{"*"}) {
		t.Error("feeds:read should match * (superadmin)")
	}
	if !ScopeMatch("articles:delete", []Scope{"*"}) {
		t.Error("articles:delete should match * (superadmin)")
	}
}

func TestScopeMatch_NoMatch(t *testing.T) {
	if ScopeMatch("feeds:read", []Scope{"articles:read"}) {
		t.Error("feeds:read should NOT match articles:read")
	}
	if ScopeMatch("feeds:write", []Scope{"feeds:read"}) {
		t.Error("feeds:write should NOT match feeds:read")
	}
}

func TestScopeMatch_ResourceSpecific(t *testing.T) {
	if !ScopeMatch("feeds:018f3a6e:read", []Scope{"feeds:018f3a6e:read"}) {
		t.Error("explicit resource-specific scope should match")
	}
	if !ScopeMatch("feeds:018f3a6e:read", []Scope{"feeds:*:read"}) {
		t.Error("feeds:*:read should match feeds:<id>:read")
	}
	if ScopeMatch("feeds:018f3a6e:read", []Scope{"feeds:otherid:read"}) {
		t.Error("different resource ID should not match")
	}
}

func TestScopeMatch_MultipleGranted(t *testing.T) {
	granted := []Scope{"feeds:read", "articles:write"}
	if !ScopeMatch("feeds:read", granted) {
		t.Error("feeds:read should be covered")
	}
	if !ScopeMatch("articles:write", granted) {
		t.Error("articles:write should be covered")
	}
	if ScopeMatch("feeds:write", granted) {
		t.Error("feeds:write should NOT be covered")
	}
}

func TestScopeMatch_EmptyGranted(t *testing.T) {
	if ScopeMatch("feeds:read", []Scope{}) {
		t.Error("nothing should match empty granted scopes")
	}
}