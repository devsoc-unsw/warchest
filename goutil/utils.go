package goutils

import "fmt"

// Assert panics if the condition is FALSE.
func Assert(condition bool, message string, args ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+message, args...))
	}
}

// Lowkey gonna make a result type
type Result[T any] struct {
	Val T
	Err error
}

func (r Result[T]) Unwrap() T {
	if r.Err != nil {
		panic(r.Err)
	}
	return r.Val
}

func Err[T any](err error) Result[T] {
	return Result[T]{
		Err: err,
	}
}

func Ok[T any](val T) Result[T] {
	return Result[T]{
		Val: val,
		Err: nil,
	}
}
