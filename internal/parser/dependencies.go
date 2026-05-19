package parser

type Dependency struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}
