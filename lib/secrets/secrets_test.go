package secrets

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPasswordGenerator(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name    string
		options DefaultPasswordOptions
		expLen  int
		exp     *regexp.Regexp
	}{
		{
			name: "all explicit", expLen: 12, exp: regexp.MustCompile("^[a-z]+$"),
			options: DefaultPasswordOptions{
				Len: intPtr(12), Digits: intPtr(0), Symbols: intPtr(0), AllowUppercase: boolPtr(false), AllowRepeat: boolPtr(true),
			},
		},
		{name: "defaults", expLen: 64, exp: nil},
		{
			name: "lowercase alpha numeric only", expLen: 64, exp: regexp.MustCompile("^[a-z]+$"),
			options: DefaultPasswordOptions{
				AllowUppercase: boolPtr(false), Digits: intPtr(0), Symbols: intPtr(0),
			},
		},
		{
			name: "alpha numeric only", expLen: 64, exp: regexp.MustCompile("^[A-Za-z]+$"),
			options: DefaultPasswordOptions{
				Digits: intPtr(0), Symbols: intPtr(0),
			},
		},
		{
			name: "alpha numeric and digits only", expLen: 64, exp: regexp.MustCompile("^[A-Za-z\\d]+$"),
			options: DefaultPasswordOptions{
				Digits: intPtr(5), Symbols: intPtr(0),
			},
		},
		{
			name: "alpha numeric only shorter Length", expLen: 32, exp: regexp.MustCompile("^[a-zA-Z]+$"),
			options: DefaultPasswordOptions{
				Len: intPtr(32), Digits: intPtr(0), Symbols: intPtr(0),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			pgen := NewDefaultPasswordGenerator(tt.options)
			password, err := pgen.Generate()
			if !assert.NoError(err) {
				return
			}
			assert.Len(password, tt.expLen)
			if tt.exp != nil {
				assert.Regexp(tt.exp, password)
			}
		})
	}
}
