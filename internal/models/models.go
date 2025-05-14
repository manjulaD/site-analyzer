package models

type Link struct {
	Href       string
	Internal   bool
	Accessible bool
}

type AnalysisResult struct {
	Title        string
	HTMLVersion  string
	Headings     map[string]int
	Internal     int
	External     int
	Inaccessible int
	LoginForm    bool
}
