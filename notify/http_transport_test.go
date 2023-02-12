package notify

import (
	"github.com/pkg/errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: Set headers must use canonical MIME Header style (eg. strict CamelCase)
func Test_validateRequestHeaders(t *testing.T) {
	tests := []struct {
		name   string
		h      http.Header
		expect error
	}{
		{
			name: "1 valid request",
			h: http.Header{
				"X-Ns-Tenantid": []string{"tenantID"},
				"X-Ns-Service":  []string{"service"},
			},
			expect: nil,
		},
		{
			name: "2 missing tenant id",
			h: http.Header{
				"X-Ns-Service": []string{"service"},
			},
			expect: errors.Wrap(ErrRequestHeaderMissingParams, "X-NS-TENANTID"),
		},
		{
			name: "3 missing service",
			h: http.Header{
				"X-Ns-Tenantid": []string{"tenantID"},
			},
			expect: errors.Wrap(ErrRequestHeaderMissingParams, "X-NS-SERVICE"),
		},
		{
			name:   "4 missing both",
			h:      http.Header{},
			expect: errors.Wrap(errors.Wrap(ErrRequestHeaderMissingParams, "X-NS-TENANTID"), "X-NS-SERVICE"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequestHeaders(tt.h)
			if tt.expect == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, tt.expect, err.Error())
			}
		})
	}
}

func Test_ValidateRequestBody(t *testing.T) {
	tests := []struct {
		name   string
		rr     Request
		expect error
	}{

		{
			name: "1 valid request",
			rr: Request{
				Type:        "http",
				HTTPRequest: "get",
			},
			expect: nil,
		},

		{
			name: "2 missing type",
			rr: Request{
				Type:        "",
				HTTPRequest: "get",
			},
			expect: errors.Wrap(ErrRequestBodyMissingParams, "type"),
		},
		{
			name: "3 missing httprequest",
			rr: Request{
				Type:        "http",
				HTTPRequest: "",
			},
			expect: errors.Wrap(ErrRequestBodyMissingParams, "http_request"),
		},
		{
			name: "4 missing both",
			rr: Request{
				Type:        "",
				HTTPRequest: "",
			},
			expect: errors.Wrap(errors.Wrap(ErrRequestBodyMissingParams, "http_request"), "type"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequestBody(tt.rr)
			if tt.expect == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, tt.expect, err.Error())
			}
		})
	}
}
