package sherpa

// Documentation object, to be returned by a Sherpa API "_docs" function.
type Docs struct {
	Title     string          `json:"title"`
	Text      string          `json:"text"`
	Functions []*FunctionDocs `json:"functions"`
	Sections  []*Docs         `json:"sections"`
}

// Documentation for a single function Name.
// Text should be in markdown. The first line should be a synopsis showing parameters including types, and the return types.
type FunctionDocs struct {
	Name string `json:"name"`
	Text string `json:"text"`
}
