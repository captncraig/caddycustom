package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mholt/caddy/middleware"
)

func (e *statsModule) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if e.statsPath != "" && middleware.Path(r.URL.Path).Matches(e.statsPath) {
		return statsHandler(w, r)
	}
	start := time.Now()
	code, err := e.next.ServeHTTP(w, r)
	duration := time.Now().Sub(start) / time.Microsecond
	path := e.pathName(r.URL.Path, r.Method)

	//every datapoint gets tagged with server and path. A few get some extra.
	tags := func(extra ...string) M {
		m := M{"path": path, "server": e.serverName}
		if len(extra)%2 == 0 {
			for i := 1; i < len(extra); i += 2 {
				m[extra[i-1]] = extra[i]
			}
		}
		return m
	}
	Statistics.Add("status_codes", tags("code", fmt.Sprint(code)), 1)
	if err != nil {
		Statistics.Add("errors", tags(), 1)
	} else {
		Statistics.Add("errors", tags(), 0)
	}
	Statistics.Add("requests", tags(), 1)
	Statistics.Sample("response_time", tags(), int64(duration))
	return code, err
}

type statsWork struct {
	statusCode int
	server     string
	path       string
	isErr      bool
	duration   int64
}

func (e *statsModule) pathName(url string, method string) string {
	for _, pth := range e.paths {
		if middleware.Path(url).Matches(pth.path) {
			if pth.methods == nil {
				return pth.name
			}
			for _, m := range pth.methods {
				if m == method {
					return pth.name
				}
			}
		}
	}
	return "/"
}

// marshal all known stats to json
func statsHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	dat, err := json.MarshalIndent(Statistics, "", "  ")
	if err != nil {
		return 0, err
	}
	if _, err = w.Write(dat); err != nil {
		return 500, err
	}
	return 200, nil
}
