package ready

import (
	"net"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("ready", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	addr, err := parse(c)
	if err != nil {
		return plugin.Error("ready", err)
	}

	rd := &ready{Addr: addr, plugins: new(list)}

	c.OnStartup(func() error {
		// Each plugin in this server block will (if they support it) report readiness.
		plugins := dnsserver.GetConfig(c).Handlers()
		for _, p := range plugins {
			if r, ok := p.(Readiness); ok {
				rd.plugins.Append(r, p.Name())
			}
		}
		return nil
	})

	c.OnStartup(rd.onStartup)
	c.OnRestart(rd.onRestart)
	c.OnFinalShutdown(rd.onFinalShutdown)

	return nil
}

func parse(c *caddy.Controller) (string, error) {
	addr := ""
	i := 0
	for c.Next() {
		if i > 0 {
			return "", plugin.ErrOnce
		}
		i++
		args := c.RemainingArgs()

		switch len(args) {
		case 0:
		case 1:
			addr = args[0]
			if _, _, e := net.SplitHostPort(addr); e != nil {
				return "", e
			}
		default:
			return "", c.ArgErr()
		}
	}
	return addr, nil
}
