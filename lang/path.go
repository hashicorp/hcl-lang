package lang

type Path struct {
	Path       string
	LanguageID string
}

func (path Path) Equals(p Path) bool {
	return path.Path == p.Path && path.LanguageID == p.LanguageID
}
