# Metrics

This module enables prometheus metrics for Caddy v2 (with custom-labels through HTTP Response Headers).

## Use

You'll need to put this module early in the chain, so that the duration histogram actually makes sense. I've put it at number 0.
In your `Caddyfile`, at global section:

```
{
    debug
    order prometheus first
}
```

For each virtual host that you want to see metrics for, these are the (optional) parameters that can be used:

  - **use_caddy_addr** - causes metrics to be exposed at the same address:port as Caddy itself. This can not be specified at the same time as **address**.
  - **address** - the address where the metrics are exposed, the default is `localhost:9180`
  - **path** - the path to serve collected metrics from, the default is `/metrics`
  - **hostname** - the `host` parameter that can be found in the exported metrics, this defaults to the label specified for the server block
  - **label** - Custom label to add on all metrics.
    This directive can be used multiple times.  
    You should specify a label name and a value.  
    The value is a placeholder {>HEADER-NAME} and can be used to extract value from response headers.  
    Usage: `label route_name {>X-Route-Name}`

## Sample Config
```
{
    debug
    order prometheus first
}

localhost:80 {
	@path1 {
        path /path1
    }

	@path2 {
        path /path2
    }

    handle @path1 {
        respond "Hello path1" 200

        header {
            "X-Route-Name" "/path1"
        }
    }

    handle @path2 {
        header {
                "X-Route-Name" "/path2"
        }

        respond "Hello path2" 200
    }

    prometheus {
        address 0.0.0.0:2081
        path    /metrics
        label route_name {>Server}
        label route_name2 {>X-Route-Name}
    }

    #metrics /metrics
}
```

## Metrics

The following metrics are exported:

* caddy_http_request_count_total{host, family, proto, ...labels}
* caddy_http_request_duration_seconds{host, family, proto, ...labels}
* caddy_http_response_latency_seconds{host, family, proto, status, ...labels}
* caddy_http_response_size_bytes{host, family, proto, status, ...labels}
* caddy_http_response_status_count_total{host, family, proto, status, ...labels}

Each metric has the following labels:

* `host` which is the hostname used for the request/response,
* `family` which is the protocol family, either 1 (IP version 4) or 2 (IP version 6),
* `proto` which is the HTTP protocol major and minor version used: 1.x or 2 signifying HTTP/1.x or HTTP/2.

The `response_*` metrics have an extra label `status` which holds the status code.
