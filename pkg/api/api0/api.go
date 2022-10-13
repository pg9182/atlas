// Package api0 implements the original master server API.
//
// External differences:
//   - Proper HTTP response codes are used (this won't break anything since existing code doesn't check them).
//   - Caching headers are supported and used where appropriate.
//   - Pdiff stuff has been removed (this was never fully implemented to begin with; see docs/PDATA.md).
//   - Error messages have been improved. Enum values remain the same for compatibility.
//   - Some rate limits (no longer necessary due to increased performance and better caching) have been removed.
//   - More HTTP methods and features are supported (e.g., HEAD, OPTIONS, Content-Encoding).
package api0

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	PdataStorage PdataStorage
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Atlas")

	switch {
	case strings.HasPrefix(r.URL.Path, "/player/"):
		// TODO: rate limit
		h.handlePlayer(w, r)
		return
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func respJSON(w http.ResponseWriter, status int, obj any) {
	buf, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.WriteHeader(status)
	w.Write(buf)
}

func respMaybeCompress(w http.ResponseWriter, r *http.Request, status int, buf []byte) {
	for _, e := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if t, _, _ := strings.Cut(e, ";"); strings.TrimSpace(t) == "gzip" {
			var cbuf bytes.Buffer
			gw := gzip.NewWriter(&cbuf)
			if _, err := gw.Write(buf); err != nil {
				break
			}
			if err := gw.Close(); err != nil {
				break
			}
			if cbuf.Len() < int(float64(len(buf))*0.8) {
				buf = cbuf.Bytes()
				w.Header().Set("Content-Encoding", "gzip")
			}
			break
		}
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.WriteHeader(status)
	w.Write(buf)
}
