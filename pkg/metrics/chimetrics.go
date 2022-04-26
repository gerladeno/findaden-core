package metrics

import (
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

const ErrDescriptionCtxKey = ContextKey("errDescription")

func NewPromMiddleware(host string) func(next http.Handler) http.Handler {
	ip, port := getExposedIPPort()
	c := NewHTTPIn(host, ip, port).AutoRegister()
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			rHost, rPort := splitHostPort(r.RemoteAddr)
			labels := []string{rHost, rPort, r.Method, chi.RouteContext(r.Context()).RoutePattern()}
			reqDump, _ := httputil.DumpRequest(r, true)
			c.ReqTotal.WithLabelValues(labels...).Inc()
			c.ReqBytesTotal.WithLabelValues(labels...).Add(float64(len(reqDump)))
			wrappedWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrappedWriter, r)

			labels = append(labels, strconv.Itoa(wrappedWriter.Status()))
			c.RespTotal.WithLabelValues(labels...).Inc()
			c.RespBytesTotal.WithLabelValues(labels...).Add(float64(wrappedWriter.BytesWritten()))
			c.RespTimeHist.WithLabelValues(labels...).Observe(float64(time.Since(started).Milliseconds()))
			c.RespTimeTotal.WithLabelValues(labels...).Set(float64(time.Since(started).Milliseconds()))
			if wrappedWriter.Status() < 500 {
				return
			}
			errDescription, ok := r.Context().Value(ErrDescriptionCtxKey).(string)
			if !ok {
				errDescription = "UNKNOWN"
			}
			c.ReqErrorsTotal.WithLabelValues(labels...).Inc()
			labels = append(labels, errDescription)
			c.RespErrorsTotal.WithLabelValues(labels...).Inc()
		}
		return http.HandlerFunc(fn)
	}
}

func getExposedIPPort() (string, string) {
	addresses, _ := net.InterfaceAddrs()
	ip := "UNKNOWN"
	port := "3000"
	for _, addr := range addresses {
		if addr.String() != "127.0.0.1/8" {
			ip = addr.String()[:strings.Index(addr.String(), "/")] // nolint:gocritic
			break
		}
	}
	return ip, port
}

func splitHostPort(hostPort string) (string, string) {
	var host, port string
	x := strings.Split(hostPort, ":")
	if len(x) > 1 {
		port = x[len(x)-1]
		x = x[:len(x)-1]
	}
	host = strings.Join(x, ":")
	return host, port
}

var reUUID = regexp.MustCompile(`\b[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}\b`)

func replaceUUIDs(path string) string {
	return reUUID.ReplaceAllString(path, "*")
}

var reAccount = regexp.MustCompile(`\d+$`)

func replaceAccounts(path string) string {
	return reAccount.ReplaceAllString(path, "*")
}
