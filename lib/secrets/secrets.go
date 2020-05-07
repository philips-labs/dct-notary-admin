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

func intPtr(v int) *int {
	return &v
}
func boolPtr(v bool) *bool {
	return &v
}

func NewDefaultPasswordGenerator(options DefaultPasswordOptions) *DefaultPasswordGenerator {
	if options.Len == nil {
		options.Len = intPtr(64)
	}
	if options.Digits == nil {
		options.Digits = intPtr(10)
	}
	if options.Symbols == nil {
		options.Symbols = intPtr(10)
	}
	if options.AllowUppercase == nil {
		options.AllowUppercase = boolPtr(true)
	}
	if options.AllowRepeat == nil {
		options.AllowRepeat = boolPtr(true)
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
