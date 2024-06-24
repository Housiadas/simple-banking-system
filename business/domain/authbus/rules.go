package authbus

import (
	_ "embed"
)

// These the current set of rules we have for authbus.
const (
	RuleAuthenticate   = "auth"
	RuleAny            = "rule_any"
	RuleAdminOnly      = "rule_admin_only"
	RuleUserOnly       = "rule_user_only"
	RuleAdminOrSubject = "rule_admin_or_subject"
)

// Package name of our rego code.
const (
	opaPackage string = "housi.rego"
)

// Core OPA policies.
var (
	//go:embed rego/authentication.rego
	regoAuthentication string

	//go:embed rego/authorization.rego
	regoAuthorization string
)
