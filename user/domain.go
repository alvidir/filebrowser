package user

const (
	profileFilename  = ".profile"
	profileDirectory = "/"
)

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
