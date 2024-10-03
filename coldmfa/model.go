package coldmfa

import "time"

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
	CodeId        string     `json:"codeId"`
	Name          string     `json:"name"`
	PreferredName *string    `json:"preferredName"`
	CreatedAt     time.Time  `json:"createdAt"`
	Deleted       bool       `json:"deleted"`
	DeletedAt     *time.Time `json:"deletedAt"`
}

type PasscodeResponse struct {
	Passcode     string `json:"passcode"`
	NextPasscode string `json:"nextPasscode"`
	ServerTime   int64  `json:"serverTime"`
	Period       uint   `json:"period"`
}

type BackupRequest struct {
	Password string `json:"password"`
}

type CodeBackup struct {
	BackupVersion string       `json:"backupVersion"`
	BackupItems   []BackupItem `json:"backup"`
}

type BackupItem struct {
	GroupName     string     `json:"groupName"`
	Original      *string    `json:"original"`
	CodeName      *string    `json:"codeName"`
	PreferredName *string    `json:"preferredName"`
	CreatedAt     *time.Time `json:"createdAt"`
	Deleted       *bool      `json:"deleted"`
	DeletedAt     *time.Time `json:"deletedAt"`
}
