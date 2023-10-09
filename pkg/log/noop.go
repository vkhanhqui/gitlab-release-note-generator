package log

func NewNoopLogger() Logger {
	return &noop{}
}

type noop struct{}

func (n *noop) Error(msg string, err error, args ...interface{}) {}
func (n *noop) Info(msg string, args ...interface{})             {}
func (n *noop) Debug(msg string, args ...interface{})            {}
func (n *noop) With(args ...interface{}) Logger                  { return &noop{} }
