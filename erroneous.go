package erroneous

import "gopkg.in/stack.v1"

// Wrap produces an Error from any existing error.
func Wrap(err error, ctx ...interface{}) Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(Error); ok {
		return e
	}

	return &erroneous{
		error: err,
		ctx:   validate(ctx),
		stack: stack.Trace()[1:].TrimRuntime(),
	}
}

// Error is erroneous's extensions to the error interface.
type Error interface {
	error

	// Unwrap returns the originally wrapped standard error.
	Unwrap() error

	// Context returns alternating keys and values
	// describing the Error's surrounding circumstances.
	Context() []interface{}

	// HTTPCode returns the http status code set on the error.
	HTTPCode() int

	// Value retrieves the value for a specific key in the context.
	Value(string) interface{}

	// Stack is the callstack associated with the error (or nil).
	Stack() stack.CallStack

	// WithContext adds further context key/values to a wrapped error.
	WithContext(...interface{}) Error

	// WithHTTPCode wraps the error replacing any HTTP status code on it.
	WithHTTPCode(int) Error

	// WithStack wraps the current error replacing the stored callstack.
	WithStack() Error
}

const httpCode = "httpcode"

type erroneous struct {
	error
	ctx   []interface{}
	stack stack.CallStack
}

func (err *erroneous) Context() []interface{} {
	return err.ctx
}

func (err *erroneous) HTTPCode() int {
	code, ok := err.Value(httpCode).(int)
	if !ok {
		return 500
	}
	return code
}

func (err *erroneous) Value(key string) interface{} {
	for i := 0; i+1 < len(err.ctx); i += 2 {
		if err.ctx[i] == key {
			return err.ctx[i+1]
		}
	}
	return nil
}

func (err *erroneous) Stack() stack.CallStack {
	return err.stack
}

func (err *erroneous) WithStack() Error {
	return &erroneous{
		error: err.error,
		ctx:   err.ctx,
		stack: stack.Trace()[1:].TrimRuntime(),
	}
}

func (err *erroneous) WithContext(ctx ...interface{}) Error {
	ctx = validate(ctx)
	newCtx := make([]interface{}, len(err.ctx)+len(ctx))
	n := copy(newCtx, err.ctx)
	n += copy(newCtx[n:], ctx)

	return &erroneous{
		error: err.error,
		ctx:   newCtx[:n],
		stack: err.stack,
	}
}

func (err *erroneous) WithHTTPCode(code int) Error {
	for i := 0; i < len(err.ctx); i += 2 {
		if err.ctx[i] == httpCode {
			ctx := make([]interface{}, len(err.ctx))
			copy(ctx, err.ctx)
			ctx[i+1] = code

			return &erroneous{
				error: err.error,
				ctx:   ctx,
				stack: err.stack,
			}
		}
	}
	return err.WithContext(httpCode, code)
}

func (err *erroneous) Unwrap() error {
	return err.error
}

func validate(ctx []interface{}) []interface{} {
	if len(ctx)%2 != 0 {
		return badCtx("invalid context, odd length")
	}
	for i := 0; i < len(ctx); i += 2 {
		_, ok := ctx[i].(string)
		if !ok {
			return badCtx("invalid context, even-position not a string")
		}
	}
	return ctx
}

func badCtx(msg string) []interface{} {
	return []interface{}{"ERRONEOUS_ERROR", msg}
}
