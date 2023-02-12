package httpget

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSender struct {
	mock.Mock
}

func (m *mockSender) Do(r *http.Request) (*http.Response, error) {
	args := m.Called(r)
	return args.Get(0).(*http.Response), args.Error(1)
}

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, b []byte) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func Test_send(t *testing.T) {
	url := ""

	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}

	s := &mockSender{}
	s.On("Do", mock.Anything).Return(r, nil)

	err := send(s, url)

	assert.Nil(t, err)
}

func Test_send_Should_Return_Err_When_Unable_Create_Request(t *testing.T) {
	url := ":"

	err := send(nil, url)

	assert.NotNil(t, err)

	s := err.Error()
	assert.True(t, strings.Contains(s, `parse ":"`))
}

func Test_send_Should_Return_Err_When_Request_Fails(t *testing.T) {
	url := ""

	s := &mockSender{}
	s.On("Do", mock.Anything).Return((*http.Response)(nil), errors.New("unable to send request"))

	err := send(s, url)

	assert.EqualError(t, err, "unable to send request")
}

func Test_send_Should_Return_Err_When_Response_Code_Is_Not_OK(t *testing.T) {
	url := ""

	r := &http.Response{
		StatusCode: 0,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}

	s := &mockSender{}
	s.On("Do", mock.Anything).Return(r, nil)

	err := send(s, url)

	assert.EqualError(t, err, "0 - : status code")
}

func Test_createTask(t *testing.T) {
	w := &worker{}

	f := w.createTask("url")

	assert.NotNil(t, f)
}