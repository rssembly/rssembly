package auth

// Pre-defined permission scopes for the RSSembly API.
// These are the resource:action pairs used by ScopeMatch.
// Adding a new scope here is the first step when adding a new route.
const (
	// Wildcard — superadmin access to everything.
	ScopeSuperadmin = "*"

	// Feeds
	ScopeFeedsRead   = "feeds:read"
	ScopeFeedsWrite  = "feeds:write"
	ScopeFeedsDelete = "feeds:delete"

	// Articles
	ScopeArticlesRead  = "articles:read"
	ScopeArticlesWrite = "articles:write"

	// Folders
	ScopeFoldersRead   = "folders:read"
	ScopeFoldersWrite  = "folders:write"
	ScopeFoldersDelete = "folders:delete"

	// Users
	ScopeUsersRead  = "users:read"
	ScopeUsersWrite = "users:write"

	// Settings
	ScopeSettingsRead  = "settings:read"
	ScopeSettingsWrite = "settings:write"
)
