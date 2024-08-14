package mongo

import "fmt"

var (
	errCollNil = New(120, "collection must not be nil")
)

type CustomError struct {
	StatusCode uint
	Status     string
	err        error
}

func (c CustomError) Error() string {
	if c.err == nil {
		return fmt.Sprintf("%d %s", c.StatusCode, c.Status)
	}
	return fmt.Sprintf("%d %s %v", c.StatusCode, c.Status, c.err)
}

func New(code uint, status string) CustomError {
	return CustomError{
		StatusCode: code,
		Status:     status,
	}
}
func (c CustomError) Throw(err error) CustomError {
	return CustomError{StatusCode: c.StatusCode, Status: c.Status, err: err}
}

func InternalServerError(err error) CustomError {
	return New(500, "problem reading clusters").Throw(err)
}
