package middleware

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "api context value " + k.name
}
