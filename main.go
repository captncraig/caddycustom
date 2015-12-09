package main

import (
	"github.com/abiosoft/caddy-git"
	"github.com/captncraig/caddy-cmd"
	"github.com/captncraig/caddy-stats"
	"github.com/mholt/caddy/caddy"
)

func main() {
	caddy.RegisterDirective("git", git.Setup, "shutdown")
	caddy.RegisterDirective("stats", stats.Setup, "shutdown")
	caddy.RegisterDirective("cmd", cmd.Setup, "basicauth")
	caddy.Main()
}
