// Package jsonrepair is a vendored copy of github.com/kaptinlin/jsonrepair at
// v0.2.2 (MIT License — see the LICENSE file in this directory, Copyright (c)
// 2024 KaptinLin).
//
// It is vendored rather than taken as a module dependency so that starlet
// keeps zero third-party dependencies and stays pinned to a release whose Go
// floor is 1.18/1.19; later jsonrepair releases raise the floor to go1.24.
// Only the runtime files are copied (const.go, errors.go, jsonrepair.go,
// utils.go) — the upstream test suite and its testify/yaml.v3 test
// dependencies are intentionally excluded. Do not edit these files by hand;
// re-vendor from the upstream tag to update.
package jsonrepair
