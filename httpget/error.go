package httpget

import "github.com/pkg/errors"

// ErrsErrStatusCode is raised when request does not return 200
var ErrStatusCode = errors.New("status code")
