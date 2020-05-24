package metrics

import (
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"net/http/httptest"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func TestMetrics_ServeHTTP(t *testing.T) {
	successRequest, err := http.NewRequest("GET", "http://test.com/success", nil)
	errorRequest, err := http.NewRequest("GET", "http://test.com/error", nil)
	proxyRequest, err := http.NewRequest("GET", "http://test.com/proxy", nil)
	proxyErrorRequest, err := http.NewRequest("GET", "http://test.com/proxyerror", nil)

	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		next caddyhttp.Handler
		addr string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "200 handler response",
			fields: fields{
				next: testHandler{},
				addr: "success",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: successRequest,
			},
			want:    200,
			wantErr: false,
		},
		{
			name: "500 handler response",
			fields: fields{
				next: testHandler{},
				addr: "error",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: errorRequest,
			},
			want:    500,
			wantErr: false,
		},
		{
			name: "proxy handler response",
			fields: fields{
				next: testHandler{},
				addr: "proxy",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: proxyRequest,
			},
			want:    200,
			wantErr: false,
		},
		{
			name: "proxy error handler response",
			fields: fields{
				next: testHandler{},
				addr: "proxyerror",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: proxyErrorRequest,
			},
			want:    502,
			wantErr: true,
		},
	}

	m := &Metrics{
		Addr:    tests[0].fields.addr,
		once:    sync.Once{},
		handler: promhttp.Handler(),
		Path:    "/metrics",
	}
	m.start()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.ServeHTTP(tt.args.w, tt.args.r, tests[0].fields.next)
			if (err != nil) != tt.wantErr {
				t.Errorf("Metrics.ServeHTTP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := tt.args.w.(*httptest.ResponseRecorder).Code
			if got != tt.want {
				t.Errorf("Metrics.ServeHTTP() [%v] = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsIPv6(t *testing.T) {
	cases := []struct {
		addr   string
		isIPv6 bool
	}{
		{"", false},
		{"192.0.2.42", false},
		{"192.0.2.42:5678", false},
		{"2001:db8::42", true},
		{"[2001:db8::42]:5678", true},
		{"banana", false},
		{"banana::phone", false},
	}

	for _, tc := range cases {
		res := isIPv6(tc.addr)
		if res != tc.isIPv6 {
			t.Errorf("isIPv6(%q) => %v, want %v", tc.addr, res, tc.isIPv6)
		}
	}
}

type testHandler struct{}

func (h testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	var err error

	switch r.URL.Path {
	case "/success":
		w.WriteHeader(200)
	case "/error":
		w.WriteHeader(500)
	case "/proxyerror":
		w.WriteHeader(502)
		err = errors.New("no hosts available upstream")
	}

	return err
}
