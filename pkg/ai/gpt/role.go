package gpt

type Role = string

const (
	// RoleUser represents a user role.
	RoleUser Role = "user"
	// RoleSystem represents a system role.
	RoleSystem   Role = "system"
	RoleFunction Role = "function"
)
