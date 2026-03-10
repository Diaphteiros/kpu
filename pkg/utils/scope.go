package utils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

type Scope string

const (
	SCOPE_ALL        Scope = "all"
	SCOPE_CLUSTER    Scope = "cluster"
	SCOPE_NAMESPACED Scope = "namespaced"
)

// Set implements pflag.Value
func (s *Scope) Set(val string) error {
	*s = Scope(val)
	return nil
}

// Type implements pflag.Value
func (s *Scope) Type() string {
	return "string"
}

// String implements pflag.Value
func (s *Scope) String() string {
	return string(*s)
}

var Scopes = []string{string(SCOPE_ALL), string(SCOPE_CLUSTER), string(SCOPE_NAMESPACED)}

// IncludesNamespaced returns true if the scope includes namespaced resources.
func (s *Scope) IncludesNamespaced() bool {
	return *s == SCOPE_NAMESPACED || *s == SCOPE_ALL
}

// IncludesCluster returns true if the scope includes cluster-scoped resources.
func (s *Scope) IncludesCluster() bool {
	return *s == SCOPE_CLUSTER || *s == SCOPE_ALL
}

// AddScopeFlag adds a --scope flag to the given FlagSet, binding the result to the given string variable.
func AddScopeFlag(fs *pflag.FlagSet, scopeVar *Scope) {
	*scopeVar = SCOPE_NAMESPACED
	fs.Var(scopeVar, "scope", fmt.Sprintf("Resource scope for the command. Valid scopes are [%s].", strings.Join(Scopes, ", ")))
}

// ValidateScope verifies that a valid scope was specified and exits with error otherwise.
func ValidateScope(scope string) {
	if !slices.Contains(Scopes, scope) {
		UnknownScopeFatal(scope)
	}
}

// UnknownScopeFatal prints an error message that the specified scope is not known and then exits with status 1.
func UnknownScopeFatal(scope string) {
	Fatal(1, "unknown scope '%s'", scope)
}
