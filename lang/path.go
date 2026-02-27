// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package lang

type Path struct {
	Path       string
	LanguageID string
	// If not empty, this is the file within the directory that the path points to
	// This is useful for languages like Terraform Test which have a file-based scope
	// instead of all files in a directory sharing the same scope
	File string
}

// TODO: check whether setting File for non test files might have unintended consequences
func (path Path) Equals(p Path) bool {
	return path.Path == p.Path && path.LanguageID == p.LanguageID && (path.File == "" || p.File == "" || path.File == p.File)
}
