package metrics

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		expected  *Metrics
	}{
		{`prometheus`, false, &Metrics{Addr: defaultAddr, Path: defaultPath, Labels: []extraLabel{}}},
		{`prometheus foo:123`, false, &Metrics{Addr: "foo:123", Path: defaultPath, Labels: []extraLabel{}}},
		{`prometheus foo bar`, true, nil},
		{`prometheus {
			a b
		}`, true, nil},
		{`prometheus
			prometheus`, true, nil},
		{`prometheus {
			address
		}`, true, nil},
		{`prometheus {
			Path
		}`, true, nil},
		{`prometheus {
			Hostname
		}`, true, nil},
		{`prometheus {
			address 0.0.0.0:1234
			use_caddy_addr
		}`, true, nil},
		{`prometheus {
			use_caddy_addr
			address 0.0.0.0:1234
		}`, true, nil},
		{`prometheus {
			use_caddy_addr
		}`, false, &Metrics{UseCaddyAddr: true, Addr: defaultAddr, Path: defaultPath, Labels: []extraLabel{}}},
		{`prometheus {
			Path /foo
		}`, false, &Metrics{Addr: defaultAddr, Path: "/foo", Labels: []extraLabel{}}},
		{`prometheus {
			use_caddy_addr
			Hostname example.com
		}`, false, &Metrics{UseCaddyAddr: true, Hostname: "example.com", Addr: defaultAddr, Path: defaultPath, Labels: []extraLabel{}}},
		{`prometheus {
			label version 1.2
			label route_name {<X-Route-Name}
		}`, false, &Metrics{Addr: defaultAddr, Path: defaultPath, Labels: []extraLabel{extraLabel{"version", "1.2"}, extraLabel{"route_name", "{<X-Route-Name}"}}}},
		{`prometheus {
			latency_buckets
		}`, true, nil},
		{`prometheus {
			latency_buckets 0.1 2 5 10
		}`, false, &Metrics{Addr: defaultAddr, Path: defaultPath, Labels: []extraLabel{}, latencyBuckets: []float64{0.1, 2, 5, 10}}},
		{`prometheus {
			size_buckets
		}`, true, nil},
		{`prometheus {
			size_buckets 1 5 10 50 100 1e3 10e6
		}`, false, &Metrics{Addr: defaultAddr, Path: defaultPath, Labels: []extraLabel{}, sizeBuckets: []float64{1, 5, 10, 50, 100, 1e3, 10e6}}},
	}
	for i, test := range tests {
		h := httpcaddyfile.Helper{
			//Dispenser:    caddyfile.NewDispenser(segment),
			Dispenser: caddyfile.NewTestDispenser(test.input),
			State:     make(map[string]interface{}),
		}
		m, err := parseCaddyfile(h)
		if test.shouldErr && err == nil {
			t.Errorf("Test %v: Expected error but found nil", i)
		} else if !test.shouldErr && err != nil {
			t.Errorf("Test %v: Expected no error but found error: %v", i, err)
		}
		if !reflect.DeepEqual(test.expected, m) && !reflect.DeepEqual(*test.expected, *(m.(*Metrics))) {
			t.Errorf("Test %v: Created Metrics (\n%#v\n) does not match expected (\n%#v\n)", i, m, test.expected)
		}
	}
}
