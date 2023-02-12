package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
)

type Endpoints struct {
	NotifyEndpoint endpoint.Endpoint
}

func NewEndpoints(svc Service) (ep Endpoints) {
	ep.NotifyEndpoint = makeEndpoint(svc)
	return ep
}

func makeEndpoint(svc Service) (ep endpoint.Endpoint) {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := request.(Request)

		if err := validateRequest(r); err != nil {
			return nil, errors.Wrap(ErrRequestBodyMissingParams, err.Error())
		}

		switch r.Type {
		case "httpget":
			return nil, svc.Send(ctx, r.HTTPRequest)
		default:
			return nil, ErrInvalidParameter
		}
	}
}

func validateRequest(r Request) error {
	s := strings.Builder{}

	if r.Type == "" {
		s.WriteString(", type")
	}

	if r.HTTPRequest == "" {
		s.WriteString(", http_request")
	}

	str := s.String()

	if len(str) > 0 {
		return fmt.Errorf("missing fields: %s", str[2:])
	}

	return nil
}
