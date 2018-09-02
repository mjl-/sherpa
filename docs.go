package sherpa

// Doc represents documentation about a Sherpa API, as returned by the "_docs" function.
type Doc struct {
	Title     string         `json:"title"`     // Name of an API section.
	Text      string         `json:"text"`      // Explanation of the API in markdown.
	Functions []*FunctionDoc `json:"functions"` // Documentation for each function exported in this API.
	Sections  []*Doc         `json:"sections"`  // Subsections, each with their own documentation.
	Types     []TypeDoc      `json:"types"`     // Types used in section or multiple subsections.

	// Version of sherpadoc format. The first version did not
	// have this field. Version 1 is the first sherpadoc with a
	// version field. Version 1 is the first with type information.
	// Only the top-level section should have a version. If
	// subsections have versions, they must be ignored.
	Version int `json:"version,omitempty"`
}

// FunctionDoc contains the documentation for a single function.
// Text should be in markdown. The first line should be a synopsis showing parameters including types, and the return types.
type FunctionDoc struct {
	Name   string  `json:"name"` // Name of the function.
	Text   string  `json:"text"` // Markdown, describing the function, its parameters, return types and possible errors.
	Params []Param `json:"params"`
	Return []Param `json:"return"`
}

// Param is the name and type of a function parameter or return value.
// Param is the name and type of a function parameter or return value.
// Type is an array of tokens describing the type.
// Production rules:
//
// 	basictype := "boolean" | "int" | "float" | "string"
// 	array := "[]"
// 	map := "{}"
// 	identifier := [a-zA-Z][a-zA-Z0-9]*
// 	type := "nullable"? ("any" | basictype | identifier | array type | map type)
//
// It is not possible to have inline structs in parameters. Those
// must be encoded as a named type.
type Param struct {
	Name string   `json:"name"`
	Type []string `json:"type"`
}

// TypeDoc is used as parameter or return value.
type TypeDoc struct {
	Name   string     `json:"name"`
	Fields []FieldDoc `json:"fields"`
	Text   string     `json:"text"`
}

// FieldDoc is a single field of a compound type.
// The type can reference another named type.
type FieldDoc struct {
	Name string   `json:"name"`
	Type []string `json:"type"`
	Text string   `json:"text"`
}
