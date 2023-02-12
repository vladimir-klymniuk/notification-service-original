package config

import "errors"

// ErrRequiredParameter is raised when there is no required configuraion parameter
var ErrRequiredParameter = errors.New("missing required parameter")
