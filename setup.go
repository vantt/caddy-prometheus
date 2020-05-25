package metrics

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(NewMetrics())
	httpcaddyfile.RegisterHandlerDirective("prometheus", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	m := new(Metrics)
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

const (
	defaultPath = "/metrics"
	defaultAddr = "localhost:9180"
)

var once sync.Once

// Metrics holds the prometheus configuration.
type Metrics struct {
	Addr           string `json:"addr,omitempty"`
	UseCaddyAddr   bool   `json:"use_caddy_addr,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Path           string `json:"path,omitempty"`
	extraLabels    []extraLabel
	latencyBuckets []float64
	sizeBuckets    []float64
	// subsystem?
	once    sync.Once
	handler http.Handler
	logger  *zap.Logger
}

type extraLabel struct {
	name  string
	value string
}

// Println implements promhttp.Logger interface, so `*Metrics` can be used as `ErrorLog`
func (m *Metrics) Println(v ...interface{}) {
	m.logger.Sugar().Error(v...)
}

// Provision initialize the metrics plugin
func (m *Metrics) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger(m)
	m.handler = promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
		ErrorLog:      m,
	})
	return m.start()
}

func (m *Metrics) Cleanup() error {
	// TODO Stop http.handle gorountine?
	return m.logger.Sync()
}

// UnmarshalCaddyfile: ?
func (m *Metrics) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		//if metrics != nil {
		//	return nil, d.Err("prometheus: can only have one metrics module per server")
		//}
		args := d.RemainingArgs()

		switch len(args) {
		case 0:
		case 1:
			m.Addr = args[0]
		default:
			return d.ArgErr()
		}
		addrSet := false
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "path":
				args = d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				m.Path = args[0]
			case "address":
				if m.UseCaddyAddr {
					return d.Err("prometheus: address and use_caddy_addr options may not be used together")
				}
				args = d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				m.Addr = args[0]
				addrSet = true
			case "hostname":
				args = d.RemainingArgs()
				if len(args) != 1 {
					return d.ArgErr()
				}
				m.Hostname = args[0]
			case "use_caddy_addr":
				if addrSet {
					return d.Err("prometheus: address and use_caddy_addr options may not be used together")
				}
				m.UseCaddyAddr = true
			case "label":
				args = d.RemainingArgs()
				if len(args) != 2 {
					return d.ArgErr()
				}

				labelName := strings.TrimSpace(args[0])
				labelValuePlaceholder := args[1]

				m.extraLabels = append(m.extraLabels, extraLabel{name: labelName, value: labelValuePlaceholder})
			case "latency_buckets":
				args = d.RemainingArgs()
				if len(args) < 1 {
					return d.Err("prometheus: must specify 1 or more latency buckets")
				}
				m.latencyBuckets = make([]float64, len(args))
				for i, v := range args {
					b, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return d.Errf("prometheus: invalid bucket %q - must be a number", v)
					}
					m.latencyBuckets[i] = b
				}
			case "size_buckets":
				args = d.RemainingArgs()
				if len(args) < 1 {
					return d.Err("prometheus: must specify 1 or more size buckets")
				}
				m.sizeBuckets = make([]float64, len(args))
				for i, v := range args {
					b, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return d.Errf("prometheus: invalid bucket %q - must be a number", v)
					}
					m.sizeBuckets[i] = b
				}
			default:
				return d.Errf("prometheus: unknown item: %s", d.Val())
			}
		}
	}
	return nil
}

// CaddyModule provides module information to Caddy
func (Metrics) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.prometheus",
		New: func() caddy.Module { // This only creates an empty metrics plugin instance
			return NewMetrics()
		},
	}
}

// NewMetrics creates an empty Metrics with default settings
func NewMetrics() *Metrics {
	return &Metrics{
		Path:        defaultPath,
		Addr:        defaultAddr,
		extraLabels: []extraLabel{},
	}
}

// start registers Prometheus routes and (optionally) starts an HTTP server to handle client scraps
func (m *Metrics) start() error {
	m.once.Do(func() {
		m.define("")

		prometheus.MustRegister(requestCount)
		prometheus.MustRegister(requestDuration)
		prometheus.MustRegister(responseLatency)
		prometheus.MustRegister(responseSize)
		prometheus.MustRegister(responseStatus)

		if !m.UseCaddyAddr {
			http.Handle(m.Path, m.handler)
			go func() {
				err := http.ListenAndServe(m.Addr, nil)
				if err != nil {
					m.logger.Error("start prometheus handler", zap.Error(err))
				}
			}()
		}
	})
	return nil
}

func (m *Metrics) extraLabelNames() []string {
	names := make([]string, 0, len(m.extraLabels))

	for _, label := range m.extraLabels {
		names = append(names, label.name)
	}

	return names
}
