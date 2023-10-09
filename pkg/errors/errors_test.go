package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachErrorCode_OnlyAttach_When_EvaluateFunc_Returns_True(t *testing.T) {
	tcs := []struct {
		name   string
		wrapFn func(error, string, EvaluateFunc) error
	}{
		{"NotFound",
			WithNotFoundE,
		},
		{"Invalid",
			func(err error, code string, ef EvaluateFunc) error { return WithInvalidE(err, code, ef) },
		},
		{"Temporary",
			WithTemporaryE,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := New(tc.name)
			nerr := tc.wrapFn(err, "code", func(err error) bool { return true })
			assert.Equal(t, "code", ErrorCode(nerr))

			terr := tc.wrapFn(err, "code", func(err error) bool { return false })
			assert.Equal(t, "", ErrorCode(terr))
		})
	}

}
