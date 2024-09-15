package coldmfa

type ApiError struct {
	Error string `json:"error"`
}

type CodeGroup struct {
	GroupId string        `json:"group_id"`
	Name    string        `json:"name"`
	Codes   []CodeSummary `json:"codes"`
}

type CreateCode struct {
	Original string `json:"original"`
}

type CodeSummary struct {
	CodeId        string  `json:"code_id"`
	Name          string  `json:"name"`
	PreferredName *string `json:"preferred_name"`
}
