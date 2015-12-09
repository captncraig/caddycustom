package stats

import (
	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
	"time"
)

type statsModule struct {
	next       middleware.Handler
	statsPath  string
	serverName string
	paths      []pathMatch
}

type pathMatch struct {
	path    string
	name    string
	methods []string
}

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	module, err := parse(c)
	if err != nil {
		return nil, err
	}
	module.serverName = c.Address()

	return func(next middleware.Handler) middleware.Handler {
		module.next = next
		return module
	}, nil
}

func parse(c *setup.Controller) (*statsModule, error) {
	var module *statsModule

	var err error
	for c.Next() {
		if module != nil {
			return nil, c.Err("Can only create one stats module")
		}
		module = &statsModule{}
		args := c.RemainingArgs()

		switch len(args) {
		case 0:
		case 1:
			module.statsPath = args[0]
		default:
			return nil, c.ArgErr()
		}

		for c.NextBlock() {
			switch c.Val() {
			case "path":
				//path /foo
				args = c.RemainingArgs()
				if len(args) < 2 {
					return nil, c.ArgErr()
				}
				pth := pathMatch{
					path: args[0],
					name: args[1],
				}
				for _, meth := range args[2:] {
					pth.methods = append(pth.methods, meth)
				}
				module.paths = append(module.paths, pth)
			case "send":
				// send dbtype args
				args = c.RemainingArgs()
				l := len(args)
				if l < 1 {
					return nil, c.ArgErr()
				}
				switch args[0] {
				case "influx":
					// influx server db interval uname password
					if l < 3 {
						return nil, c.ArgErr()
					}
					server := args[1]
					db := args[2]
					interval := 15 * time.Second
					uname, password := "", ""
					if l >= 4 {
						interval, err = time.ParseDuration(args[3])
						if err != nil {
							return nil, err
						}
					}
					if l == 5 || l > 6 {
						return nil, c.ArgErr()
					}
					if l == 6 {
						uname, password = args[4], args[5]
					}
					go sendInflux(server, db, uname, password, interval)
				default:
					return nil, c.Errf("Unknown send db: %s", args[0])
				}
			default:
				return nil, c.Errf("Unknown stats config item: %s", c.Val())
			}

		}
	}
	return module, err
}
