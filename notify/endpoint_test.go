package notify

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_validateRequest(t *testing.T) {
	tests := []struct {
		name   string
		r      Request
		expect error
	}{
		{
			name: "1 valid request",
			r: Request{
				Type:        "httpget",
				HTTPRequest: "wow",
			},
			expect: nil,
		},
		{
			name: "2 missing type",
			r: Request{
				HTTPRequest: "wow",
			},
			expect: errors.New("missing fields: type"),
		},
		{
			name: "3 missing http_request",
			r: Request{
				Type: "httpget",
			},
			expect: errors.New("missing fields: http_request"),
		},
		{
			name:   "4 missing both",
			r:      Request{},
			expect: errors.New("missing fields: type, http_request"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequest(tt.r)
			if tt.expect == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, tt.expect, err.Error())
			}
		})
	}
}
