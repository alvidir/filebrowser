package event

type FileEventPayload struct {
	Issuer   string `json:"issuer"`
	UserID   int32  `json:"user_id"`
	AppID    string `json:"app_id"`
	FileName string `json:"file_name"`
	FileID   string `json:"file_id"`
	Kind     string `json:"kind"`
}

type UserProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserEventPayload struct {
	UserID int32  `json:"id"`
	Kind   string `json:"kind"`
	UserProfile
}
