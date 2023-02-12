# NOTIFICATION-SERVICE

Notifiation Service is a service that will take care of the communication, to retry x time and fail ultimately alerting us, that we need to manually act, on it.

Notification Service receive in /notify a POST message like :

```
curl -X POST /notify -d '{
    "type": "httpget",
    "http_request": "http://myburl/..../"
}'
```
Headers:
```
"X-NS-TENANTID" = "delivery"
"X-NS-SERVICE"  = "ssp"
```
When it receives that call it will send it on kafka (in a Topic named TenantId+"-"+ServiceName example delivery-dsp)
And then a worker consuming that same topic will retry x time (configured).

If for some reasons it fails, it will send to another topic (topic named TenantId+"-"+ServiceName+"-error" like delivery-dsp-error).
With the http_request and the reason it failed (on json)

Configuration looks like :

```
[Kafka]
Servers

[Service.Name] # It's a toml table
TenantId (mandatory string example delivery)
RetryTime (optional by default = 3)
```
