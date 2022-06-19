package internalerrors

import "errors"

// ErrNotFound will be used  when a resources is missing.
var ErrNotFound = errors.New("resources does not exists")
