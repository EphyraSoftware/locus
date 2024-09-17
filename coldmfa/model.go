package coldmfa

type ApiError struct {
	Error string `json:"error"`
}

type CodeGroup struct {
	GroupId string        `json:"groupId"`
	Name    string        `json:"name"`
	Codes   []CodeSummary `json:"codes"`
}

type CreateCode struct {
	Original string `json:"original"`
}

type CodeSummary struct {
	CodeId        string  `json:"codeId"`
	Name          string  `json:"name"`
	PreferredName *string `json:"preferredName"`
	CreatedAt     int64   `json:"createdAt"`
}

type PasscodeResponse struct {
	Passcode     string `json:"passcode"`
	NextPasscode string `json:"nextPasscode"`
	ServerTime   int64  `json:"serverTime"`
	Period       uint   `json:"period"`
}
