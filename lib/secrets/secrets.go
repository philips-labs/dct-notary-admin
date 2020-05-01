package secrets

type PasswordGenerator interface {
	Generate() (string, error)
}
