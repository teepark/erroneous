package erroneous

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/stack.v1"
)

func TestNoopOnNil(t *testing.T) {
	e := Wrap(nil)
	if e != nil {
		t.Fatalf("Wrap(nil) produced non-nil: %+v", e)
	}
}

func TestWrapErrProducesErr(t *testing.T) {
	err := errors.New("failure of some sort or another")
	var e Error
	e = Wrap(err)
	if e == nil {
		t.Fatalf("Wrap() produced nil from valid err: %+v", e)
	}
}

func TestWrapDoesntDoubleWrap(t *testing.T) {
	once := Wrap(errors.New("ExplosionError"))
	twice := Wrap(once)
	if once != twice {
		t.Fatal("second wrap wasn't a noop")
	}
}

func TestWrapCanAddCtx(t *testing.T) {
	err := errors.New("your error here")
	e := Wrap(err, "foo", "bar")
	if len(e.Context()) != 2 {
		t.Fatal("wrong/no context added by Wrap()")
	}
	if e.Context()[0] != "foo" {
		t.Fatal("wrong key in context!")
	}
	if e.Context()[1] != "bar" {
		t.Fatal("wrong value in context!")
	}
}

func TestCtxValues(t *testing.T) {
	err := Wrap(errors.New("errormethis"), "foo", "bar")
	if err.Value("foo") != "bar" {
		t.Fatalf("couldn't get correct value, got: %q", err.Value("foo"))
	}
}

func TestInvalidCtxOdd(t *testing.T) {
	err := Wrap(errors.New("anerr"), "keyonly")
	ctx := err.Context()
	if len(ctx) != 2 {
		t.Fatal("got an invalid context out")
	}
	if ctx[0] != "ERRONEOUS_ERROR" {
		t.Fatalf("wrong key: %q", ctx[0])
	}
}

func TestInvalidCtxStringKey(t *testing.T) {
	err := Wrap(errors.New("anerr"), 14, "key wasn't a string")
	ctx := err.Context()
	if ctx[0] != "ERRONEOUS_ERROR" {
		t.Fatalf("wrong ctx key: %v", ctx[0])
	}
}

func TestProducesStack(t *testing.T) {
	err := Wrap(errors.New("error here"))
	st := err.Stack()
	if st == nil {
		t.Fatal("got a nil stack!")
	}
}

func TestProducesStackFromWrapSite(t *testing.T) {
	// important for this test to work: the stack.Caller()
	// call must be on the line immediately before the Wrap().
	err := Wrap(errors.New("record me"))
	pass, e := stackIsFromPreviousLine(err.Stack())
	if e != nil {
		t.Fatal(err)
	}
	if !pass {
		t.Fatal("wrong top location in stack")
	}
}

func TestStackTrimsRuntime(t *testing.T) {
	st := Wrap(errors.New("error here")).Stack()
	if strings.Contains(fmt.Sprintf("%+s", st), "runtime") {
		t.Fatalf("'runtime' found in stack: %+s", st)
	}
}

func TestWithStack(t *testing.T) {
	err := Wrap(errors.New("errOR"))

	err = err.WithStack()
	pass, e := stackIsFromPreviousLine(err.Stack())
	if e != nil {
		t.Fatal(e)
	}
	if !pass {
		t.Fatal("wrong top location in stack")
	}
}

func TestUnwrapped(t *testing.T) {
	original := errors.New("I'm an error")
	err := Wrap(original)

	if err.Unwrap() != original {
		t.Fatal("wrong unwrapped result")
	}
}

func TestWithContext(t *testing.T) {
	err := Wrap(errors.New("I'm an error, no context yet"), "old", "context")
	err2 := err.WithContext("foo", "bar", "spam", "eggs")

	if err2.Value("old") != "context" {
		t.Fatal("new error didn't inherit context")
	}

	if err2.Value("foo") != "bar" {
		t.Fatal("extra context didn't get stored")
	}
	if err2.Value("spam") != "eggs" {
		t.Fatal("extra context missing second new value")
	}
}

func TestHTTPCode(t *testing.T) {
	err := Wrap(errors.New("anerr"))
	if err.HTTPCode() != 500 {
		t.Fatal("didn't read implicit 500")
	}

	err = Wrap(errors.New("anothererr"), httpCode, 404)
	if err.HTTPCode() != 404 {
		t.Fatal("didn't read explicitly set 404 http code")
	}
}

func TestWithHTTPCode(t *testing.T) {
	err := Wrap(errors.New("anerr"))
	if err.HTTPCode() != 500 {
		t.Fatal("didn't read implicit 500")
	}

	err = err.WithHTTPCode(404)
	if err.HTTPCode() != 404 {
		t.Fatal("didn't read WithHTTPCode-set 404")
	}

	err = err.WithHTTPCode(409)
	if err.HTTPCode() != 409 {
		t.Fatal("didn't read WithHTTPCode-overridden 409")
	}
}

func stackIsFromPreviousLine(st stack.CallStack) (bool, error) {
	// for this to be accurate, you must call this function on the line
	// immediately following the expected generation of the stack you pass it.
	herestr := fmt.Sprintf("%d", stack.Caller(1))
	here, err := strconv.Atoi(herestr)
	if err != nil {
		return false, err
	}

	sitestr := fmt.Sprintf("%d", st[0])
	site, err := strconv.Atoi(sitestr)
	if err != nil {
		return false, err
	}

	return here == site+1, nil
}
