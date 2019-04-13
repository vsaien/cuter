package baseerror

type CodeError struct {
	code int
	desc string
}

func (err *CodeError) Error() string {
	return err.desc
}

func (err *CodeError) Code() int {
	return err.code
}

func NewCodeError(code int, desc string) *CodeError {
	return &CodeError{
		code: code,
		desc: desc,
	}
}

func IsCodeError(err error) bool {
	switch err.(type) {
	case *CodeError:
		return true
	}
	return false
}

func FromError(err error) (codeErr *CodeError, ok bool) {
	if se, ok := err.(*CodeError); ok {
		return se, true
	}
	return nil, false
}

func ToCodeError(err error) *CodeError {
	if IsCodeError(err) {
		return err.(*CodeError)
	}
	return ErrDefault
}
