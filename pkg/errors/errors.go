package errors

import (
	"errors"
	"fmt"
	"io"
)

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// This actually calls errors.As() of the standard package.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// WithCode annotates err with a code.
func WithCode(err error, code string) error {
	return &withCode{cause: err, code: code, stack: callers()}
}

// IsNotFound returns true if err is a NotFound error.
func IsNotFound(err error) bool {
	ne, ok := err.(exists)
	return ok && ne.NotFound()
}

// WithNotFound annotates err with NotFound behavior.
func WithNotFound(err error, code string) error {
	if err == nil {
		return nil
	}
	return &withNotFound{withCode{cause: err, code: code, stack: callers()}, nil}
}

// WithNotFoundE annotates err with NotFound behavior by given custom evaluator.
func WithNotFoundE(err error, code string, ef EvaluateFunc) error {
	if err == nil {
		return nil
	}

	nerr := &withNotFound{withCode{cause: err, stack: callers()}, ef}
	if ef(err) {
		nerr.code = code
	}

	return nerr
}

// IsInvalid returns true if err is a Validation error.
func IsInvalid(err error) bool {
	ve, ok := err.(validation)
	return ok && ve.Invalid()
}

// WithInvalid annotates err with Invalid behavior.
func WithInvalid(err error, code string, i18nData ...interface{}) error {
	if err == nil {
		return nil
	}
	return &withInvalid{
		withCode{
			cause:    err,
			code:     code,
			stack:    callers(),
			i18nData: i18nData,
		},
		nil,
	}
}

// WithInvalidE annotates err with NotFound behavior by given custom evaluator.
func WithInvalidE(err error, code string, ef EvaluateFunc, i18nData ...interface{}) error {
	if err == nil {
		return nil
	}

	ie := &withInvalid{
		withCode{
			cause:    err,
			stack:    callers(),
			i18nData: i18nData,
		},
		ef,
	}
	if ef(err) {
		ie.code = code
	}

	return ie
}

// IsTemporary returns true if err is temporary, usually used in retry context.
func IsTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}

// WithTemporary annotates err with Temporary behavior.
func WithTemporary(err error, code string) error {
	if err == nil {
		return nil
	}
	return &withTemporary{withCode{cause: err, code: code, stack: callers()}, nil}
}

// WithTemporaryE annotates err with Temporary behavior by given custom evaluator.
func WithTemporaryE(err error, code string, ef EvaluateFunc) error {
	if err == nil {
		return nil
	}

	te := &withTemporary{withCode{cause: err, stack: callers()}, ef}
	if ef(err) {
		te.code = code
	}

	return te
}

// IsTimeout returns true if err is related to a timeout.
func IsTimeout(err error) bool {
	te, ok := err.(timeout)
	return ok && te.Timeout()
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}

	type errorCode interface {
		Code() string
	}
	ec, ok := err.(errorCode)
	if !ok {
		return "unknown"
	}
	return ec.Code()
}

func I18nData(err error) []string {
	type i18nData interface {
		I18nData() []interface{}
	}
	i18nErr, ok := err.(i18nData)
	if !ok {
		return []string{}
	}

	data := i18nErr.I18nData()
	strs := make([]string, len(data))
	for i, v := range data {
		strs[i] = fmt.Sprintf("%v", v)
	}

	return strs
}

type exists interface {
	NotFound() bool
}

type validation interface {
	Invalid() bool
}

type temporary interface {
	Temporary() bool
}

type timeout interface {
	Timeout() bool
}

type withCode struct {
	code     string
	i18nData []interface{}
	cause    error
	*stack
}

func (e *withCode) Cause() error { return e.cause }

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *withCode) Unwrap() error { return e.cause }

func (e *withCode) Error() string {
	return e.cause.Error()
}

func (e *withCode) Code() string {
	return e.code
}

func (e *withCode) I18nData() []interface{} {
	return e.i18nData
}

func (e *withCode) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.cause.Error())
			e.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.cause.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.cause.Error())
	}
}

type EvaluateFunc func(error) bool

type withNotFound struct {
	withCode
	ef EvaluateFunc
}

func (e *withNotFound) NotFound() bool {
	return e.ef == nil || e.ef(e.cause)
}

type withInvalid struct {
	withCode
	ef EvaluateFunc
}

func (e *withInvalid) Invalid() bool {
	return e.ef == nil || e.ef(e.cause)
}

type withTemporary struct {
	withCode
	ef EvaluateFunc
}

func (e *withTemporary) Temporary() bool {
	return e.ef == nil || e.ef(e.cause)
}
