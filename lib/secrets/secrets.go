package secrets

import (
	"github.com/sethvargo/go-password/password"
)

type PasswordGenerator interface {
	Generate() (string, error)
}

type DefaultPasswordOptions struct {
	Len            *int  `json:"length,omitempty"`
	Digits         *int  `json:"digits,omitempty"`
	Symbols        *int  `json:"symbols,omitempty"`
	AllowUppercase *bool `json:"allow_uppercase,omitempty"`
	AllowRepeat    *bool `json:"allow_repeat,omitempty"`
}

type DefaultPasswordGenerator struct {
	options DefaultPasswordOptions
}

//go:fix inline
func intPtr(v int) *int {
	return new(v)
}

//go:fix inline
func boolPtr(v bool) *bool {
	return new(v)
}

func NewDefaultPasswordGenerator(options DefaultPasswordOptions) *DefaultPasswordGenerator {
	if options.Len == nil {
		options.Len = new(64)
	}
	if options.Digits == nil {
		options.Digits = new(10)
	}
	if options.Symbols == nil {
		options.Symbols = new(10)
	}
	if options.AllowUppercase == nil {
		options.AllowUppercase = new(true)
	}
	if options.AllowRepeat == nil {
		options.AllowRepeat = new(true)
	}

	return &DefaultPasswordGenerator{
		options: options,
	}
}

func (g *DefaultPasswordGenerator) Generate() (string, error) {
	res, err := password.Generate(
		*g.options.Len,
		*g.options.Digits,
		*g.options.Symbols,
		!*g.options.AllowUppercase,
		*g.options.AllowRepeat,
	)
	return res, err
}
