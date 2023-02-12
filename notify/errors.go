package notify

import "errors"

// ErrInvalidParameter is raised when the parameters of the request are invalid
var ErrInvalidParameter = errors.New("invalid parameter")

// ErrRequestBodyMissingParams is raised when the request body is missing mandatory fields
var ErrRequestBodyMissingParams = errors.New("request body missing params")

// ErrRequestHeaderMissingParams is raised when the request header is missing mandatory fields
var ErrRequestHeaderMissingParams = errors.New("request header missing params")