package coldmfa

type ApiError struct {
	Error string `json:"error"`
}

type CodeGroup struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}
