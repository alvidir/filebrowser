package user

const (
	profilePath = ".profile"
)

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
