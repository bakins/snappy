// Package snappy ienables snappy compression/decompression for
// HTTP clients and servers
package snappy

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/golang/snappy"
)

const snappyEncoding = "snappy"

var (
	readerPool = &sync.Pool{
		New: func() interface{} {
			return &snappyReader{
				Reader: snappy.NewReader(nil),
			}
		},
	}

	writerPool = &sync.Pool{
		New: func() interface{} {
			return &responseWriter{
				writer: snappy.NewBufferedWriter(nil),
			}
		},
	}
)

// Handler wraps an http handler with snappy compression.
func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")

		if !hasSnappyEncoding(r.Header["Accept-Encoding"]...) {
			next.ServeHTTP(w, r)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			flusher = nil
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			hijacker = nil
		}

		s := writerPool.Get().(*responseWriter)
		s.ResponseWriter = w
		s.writer.Reset(w)
		s.Flusher = flusher
		s.Hijacker = hijacker

		defer func() {
			s.writer.Close()
			writerPool.Put(s)
		}()

		w.Header().Set("Content-Encoding", snappyEncoding)
		r.Header.Del("Accept-Encoding")

		next.ServeHTTP(s, r)
	})
}

type responseWriter struct {
	writer *snappy.Writer
	http.ResponseWriter
	http.Hijacker
	http.Flusher
}

func (w *responseWriter) WriteHeader(c int) {
	w.ResponseWriter.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(c)
}

func (w *responseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	h.Del("Content-Length")

	return w.writer.Write(b)
}

func (w *responseWriter) Flush() {
	_ = w.writer.Flush()

	// Flush HTTP response.
	if w.Flusher != nil {
		w.Flusher.Flush()
	}
}

// Transport wraps an http transport to add support for client side
// Snappy compression.
// if base is nil, http.DefaultTransport is used.
func Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return &transport{
		base: base,
	}
}

type transport struct {
	base http.RoundTripper
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Accept-Encoding", snappyEncoding)

	resp, err := t.base.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	if !hasSnappyEncoding(resp.Header["Content-Encoding"]...) {
		return resp, nil
	}

	reader := readerPool.Get().(*snappyReader)
	reader.ReadCloser = resp.Body
	reader.Reader.Reset(resp.Body)

	resp.Body = reader

	return resp, nil
}

type snappyReader struct {
	io.ReadCloser
	*snappy.Reader
}

func (s *snappyReader) Close() error {
	err := s.ReadCloser.Close()

	readerPool.Put(s)

	return err
}

func (s *snappyReader) Read(p []byte) (int, error) {
	return s.Reader.Read(p)
}

func hasSnappyEncoding(in ...string) bool {
	for _, v := range in {
		if v == "" {
			continue
		}

		for _, curEnc := range strings.Split(v, ",") {
			curEnc = strings.TrimSpace(curEnc)
			if curEnc == snappyEncoding {
				return true
			}
		}
	}

	return false
}
