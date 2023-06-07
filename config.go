package starlet

import (
	"go.starlark.net/resolve"
)

// EnableRecursionSupport enables recursion support in Starlark.
func EnableRecursionSupport() {
	resolve.AllowRecursion = true
}

// DisableRecursionSupport disables recursion support in Starlark.
func DisableRecursionSupport() {
	resolve.AllowRecursion = false
}

// EnableGlobalReassign enables global reassignment in Starlark.
func EnableGlobalReassign() {
	resolve.AllowGlobalReassign = true
}

// DisableGlobalReassign disables global reassignment in Starlark.
func DisableGlobalReassign() {
	resolve.AllowGlobalReassign = false
}
