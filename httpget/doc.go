// Package httpget handles retries of GET calls. Requests made to this handler
// must have the headers X-NS-TENANTID and X-NS-SERVICE, which refer to the calling
// app's tenant ID and service name, respectively. The provided URL will then be
// published to a kafka broker, which will be consumed by the worker to handle
// retrying the GET request should it ever fail.
package httpget
