package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type handler struct {
	next     middleware.Handler
	handlers map[string]string
	root     string
}

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	module, err := parse(c)
	if err != nil {
		return nil, err
	}
	module.root = c.Root
	return func(next middleware.Handler) middleware.Handler {
		module.next = next
		return module
	}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	panic("AAA")
	fmt.Println(r.URL.Path)
	for path, templatePath := range h.handlers {
		if !middleware.Path(r.URL.Path).Matches(path) {
			continue
		}
		if filepath.Ext(r.URL.Path) != ".json" {
			continue
		}
		//load json and template files
		jsonPath := filepath.Join(h.root, r.URL.Path)
		templatePath := filepath.Join(h.root, templatePath)

		jsonFile, err := ioutil.ReadFile(jsonPath)
		if err != nil {
			return 500, err
		}
		templateFile, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return 500, err
		}
		data := map[string]interface{}{}
		if err = json.Unmarshal(jsonFile, &data); err != nil {
			return 500, err
		}
		tpl, err := template.New("").Parse(string(templateFile))
		if err != nil {
			return 500, err
		}
		buf := &bytes.Buffer{}
		if err = tpl.Execute(buf, data); err != nil {
			return 500, err
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(buf.Bytes())
		return 200, nil
	}
	return h.next.ServeHTTP(w, r)
}

func parse(c *setup.Controller) (*handler, error) {
	h := &handler{handlers: map[string]string{}}
	for c.Next() {
		args := c.RemainingArgs()
		if len(args) != 2 {
			return nil, c.ArgErr()
		}
		h.handlers[args[0]] = args[1]
	}
	return h, nil
}
