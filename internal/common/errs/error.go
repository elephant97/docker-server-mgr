package errs

type Code int

const (
	Invaild Code = iota + 1
	DefaultSet
)

type ErrorDetail struct {
	Code    int
	Message error
}
