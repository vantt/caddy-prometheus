package metrics

import (
	"bytes"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/davecgh/go-spew/spew"
)

func (m Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

	// If use_caddy_addr is configured, we're hijacking `m.Path` for prometheus endpoint
	if m.UseCaddyAddr && r.URL.Path == m.Path {
		m.handler.ServeHTTP(w, r)
		return nil
	}

	hostname := m.Hostname

	if hostname == "" {
		originalHostname, err := host(r)
		if err != nil {
			hostname = "-"
		} else {
			hostname = originalHostname
		}
	}
	start := time.Now()

	// Record response to get status code and size of the reply.
	rw := caddyhttp.NewResponseRecorder(w, bytes.NewBuffer(nil), func(status int, header http.Header) bool { return false })

	// Get time to first write.
	tw := &timedResponseWriter{ResponseWriter: rw}

	err := next.ServeHTTP(tw, r)

	status := tw.ResponseWriter.(caddyhttp.ResponseRecorder).Status()
	spew.Dump(status)
	// If nothing was explicitly written, consider the request written to
	// now that it has completed.
	tw.didWrite()

	// Transparently capture the status code so as to not side effect other plugins
	stat := status
	spew.Dump(status)
	if err != nil && status == 0 {
		spew.Dump(status)
		// Some middlewares set the status to 0, but return an non nil error: map these to status 500
		stat = 500
	} else if status == 0 {
		// 'proxy' returns a status code of 0, but the actual status is available on rw.
		// Note that if 'proxy' encounters an error, it returns the appropriate status code (such as 502)
		// from ServeHTTP and is captured above with 'stat := status'.
		stat = rw.Status()
		spew.Dump(status)
	}

	fam := "1"
	if isIPv6(r.RemoteAddr) {
		fam = "2"
	}

	proto := strconv.Itoa(r.ProtoMajor)
	proto = proto + "." + strconv.Itoa(r.ProtoMinor)

	statusStr := strconv.Itoa(stat)

	replacer := caddy.NewReplacer()
	replacer.Map(headerReplacementFunc(tw.Header()))

	var extraLabelValues []string
	for _, label := range m.Labels {
		extraLabelValues = append(extraLabelValues, replacer.ReplaceAll(label.Value, ""))
	}

	requestCount.WithLabelValues(append([]string{hostname, fam, proto}, extraLabelValues...)...).Inc()
	requestDuration.WithLabelValues(append([]string{hostname, fam, proto}, extraLabelValues...)...).Observe(time.Since(start).Seconds())
	responseSize.WithLabelValues(append([]string{hostname, fam, proto, statusStr}, extraLabelValues...)...).Observe(float64(rw.Size()))
	responseStatus.WithLabelValues(append([]string{hostname, fam, proto, statusStr}, extraLabelValues...)...).Inc()
	responseLatency.WithLabelValues(append([]string{hostname, fam, proto, statusStr}, extraLabelValues...)...).Observe(tw.firstWrite.Sub(start).Seconds())

	return err
}

func headerReplacementFunc(h http.Header) caddy.ReplacerFunc {
	return func (key string) (interface{}, bool) {
		const envPrefix = ">"
		if strings.HasPrefix(key, envPrefix) {
			return h.Get(key[1:]), true
		}
		return nil, false
	}
}


func host(r *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		if !strings.Contains(r.Host, ":") {
			return strings.ToLower(r.Host), nil
		}
		return "", err
	}
	return strings.ToLower(host), nil
}

func isIPv6(addr string) bool {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		// Strip away the port.
		addr = host
	}
	ip := net.ParseIP(addr)
	return ip != nil && ip.To4() == nil
}

// A timedResponseWriter tracks the time when the first response write
// happened.
type timedResponseWriter struct {
	firstWrite time.Time
	http.ResponseWriter
}

func (w *timedResponseWriter) didWrite() {
	if w.firstWrite.IsZero() {
		w.firstWrite = time.Now()
	}
}

func (w *timedResponseWriter) Write(data []byte) (int, error) {
	w.didWrite()
	return w.ResponseWriter.Write(data)
}

func (w *timedResponseWriter) WriteHeader(statuscode int) {
	// We consider this a write as it's valid to respond to a request by
	// just setting a status code and returning.
	w.didWrite()
	w.ResponseWriter.WriteHeader(statuscode)
}
