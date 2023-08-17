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
	InputDoc  any
	OutputDoc any
	ErrorDoc  any
	Comments  []string
	QueryVars QueryVars
}
