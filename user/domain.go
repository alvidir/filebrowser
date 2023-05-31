package user

const (
	profileFilename  = ".profile"
	profileDirectory = "/"
)

type Profile struct {
	Name  string `json:"user_name"`
	Email string `json:"user_email"`
}
