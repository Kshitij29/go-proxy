# go-proxy
Proxy server in go to handle http/https connections.

## branch: http_connect_tunneling
Proxy for handling HTTP CONNECT to open up a TCP connection through the proxy. This proxy will not be able to cache, read, or modify the HTTPS connection. HTTP connections will be handled normally.
