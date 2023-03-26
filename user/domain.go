package user

const (
	profilePath = ".profile"
)

type UserProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
