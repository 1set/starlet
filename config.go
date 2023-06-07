package starlet

import (
	"go.starlark.net/resolve"
)

// EnableRecursionSupport enables recursion support in all Starlark environments.
func EnableRecursionSupport() {
	resolve.AllowRecursion = true
}

// DisableRecursionSupport disables recursion support in all Starlark environments.
func DisableRecursionSupport() {
	resolve.AllowRecursion = false
}

// EnableGlobalReassign enables global reassignment in all Starlark environments.
func EnableGlobalReassign() {
	resolve.AllowGlobalReassign = true
}

// DisableGlobalReassign disables global reassignment in all Starlark environments.
func DisableGlobalReassign() {
	resolve.AllowGlobalReassign = false
}
