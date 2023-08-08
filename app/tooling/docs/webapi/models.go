package webapi

type Field struct {
	Name     string
	Type     string
	Tag      string
	Optional bool
}

type QueryVars struct {
	Paging   []string
	FilterBy []string
	OrderBy  []string
}

type Record struct {
	Group     string
	Method    string
	Route     string
	Status    string
	InputDoc  map[string]any
	OutputDoc map[string]any
	Comments  []string
	QueryVars QueryVars
}
