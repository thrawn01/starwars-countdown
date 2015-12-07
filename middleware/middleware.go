package middleware

import (
	"github.com/justinas/alice"
	"github.com/oxtoacart/bpool"
	"gopkg.in/throttled/throttled.v2"
	"gopkg.in/throttled/throttled.v2/store/memstore"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func Timeout(handler http.Handler) http.Handler {
	// TODO: This should return JSON
	return http.TimeoutHandler(handler, 30*time.Second, "timed out")
}

func NewThrottle() alice.Constructor {
	// Optional redis memstore available for throttling across multiple servers
	store, err := memstore.New(65536)
	if err != nil {
		log.Fatal(err)
	}

	// No more than 500 requests a minute with bursts of 50
	quota := throttled.RateQuota{throttled.PerMin(500), 50}
	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		log.Fatal(err)
	}

	httpRateLimiter := throttled.HTTPRateLimiter{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Path: false},
	}
	return httpRateLimiter.RateLimit
}

func NewRequestLogger(size int) alice.Constructor {
	req := &RequestLogger{bpool.NewBufferPool(size)}
	return req.Handler
}

// Wrap the ResponseWriter so we can intercept the status code and the size of
// the response body set by upstream middleware
type WrappedResponseWriter struct {
	resp   http.ResponseWriter
	status int
	size   int
}

func (self *WrappedResponseWriter) Header() http.Header {
	return self.resp.Header()
}

func (self *WrappedResponseWriter) Write(buf []byte) (int, error) {
	self.size += len(buf)
	return self.resp.Write(buf)
}

func (self *WrappedResponseWriter) WriteHeader(status int) {
	self.status = status
	self.resp.WriteHeader(status)
}

// Writes apache style access logs using an efficient buffer pool to avoid
// garbage collection
type RequestLogger struct {
	bufferPool *bpool.BufferPool
}

func (self *RequestLogger) Handler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(originalResp http.ResponseWriter, req *http.Request) {
		buf := self.bufferPool.Get()

		// Add the client remote address
		remoteAddress, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			remoteAddress = req.RemoteAddr
		}

		buf.WriteString(remoteAddress)
		buf.WriteString(" - ")

		// Add the authenticated user (if none then '-')
		if req.URL.User != nil {
			if name := req.URL.User.Username(); name != "" {
				buf.WriteString(name)
			} else {
				buf.WriteString("-")
			}
		} else {
			buf.WriteString("-")
		}

		// Time
		buf.WriteString(" [")
		buf.WriteString(time.Now().Format("01/Jan/2015:01:01:01 -0600"))
		buf.WriteString("] ")
		// Http Verb
		buf.WriteString(" ")
		buf.WriteString(req.Method)
		// Uri
		buf.WriteString(" \"")
		buf.WriteString(req.URL.RequestURI())
		buf.WriteString("\" ")
		// Proto
		buf.WriteString(req.Proto)
		buf.WriteString(" ")

		resp := &WrappedResponseWriter{originalResp, 200, 0}
		// Call up the middleware chain
		handler.ServeHTTP(resp, req)

		// Status Code
		buf.WriteString(strconv.Itoa(resp.status))
		buf.WriteString(" ")
		// Result Size
		buf.WriteString(strconv.Itoa(resp.size))
		// TODO: Write out the log entry in a buffered non blocking manner
		log.Println(buf)
		// Put the buffer back into the pool
		self.bufferPool.Put(buf)
	})
}
