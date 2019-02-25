package ready

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/coredns/coredns/plugin/erratic"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func init() { clog.Discard() }

func TestReady(t *testing.T) {
	rd := &ready{Addr: ":0", plugins: new(list)}
	e := &erratic.Erratic{}
	rd.plugins.Append(e, "erratic")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		if err := rd.onStartup(); err != nil {
			t.Fatalf("Unable to startup the readiness server: %v", err)
		}
		wg.Done()
	}()
	wg.Wait()

	defer rd.onFinalShutdown()

	address := fmt.Sprintf("http://%s/ready", rd.ln.Addr().String())

	wg.Add(1)
	go func() {
		response, err := http.Get(address)
		if err != nil {
			t.Fatalf("Unable to query %s: %v", address, err)
		}
		if response.StatusCode != 503 {
			t.Errorf("Invalid status code: expecting %d, got %d", 503, response.StatusCode)
		}
		response.Body.Close()
		wg.Done()
	}()
	wg.Wait()

	// make it ready by giving erratic 3 queries.
	m := new(dns.Msg)
	m.SetQuestion("example.org.", dns.TypeA)
	e.ServeDNS(context.TODO(), &test.ResponseWriter{}, m)
	e.ServeDNS(context.TODO(), &test.ResponseWriter{}, m)
	e.ServeDNS(context.TODO(), &test.ResponseWriter{}, m)

	response, err := http.Get(address)
	if err != nil {
		t.Fatalf("Unable to query %s: %v", address, err)
	}
	if response.StatusCode != 200 {
		t.Errorf("Invalid status code: expecting %d, got %d", 200, response.StatusCode)
	}
	response.Body.Close()

	// make erratic not-ready by giving it more queries, this should not change the process readiness
	e.ServeDNS(context.TODO(), &test.ResponseWriter{}, m)
	e.ServeDNS(context.TODO(), &test.ResponseWriter{}, m)
	e.ServeDNS(context.TODO(), &test.ResponseWriter{}, m)

	response, err = http.Get(address)
	if err != nil {
		t.Fatalf("Unable to query %s: %v", address, err)
	}
	if response.StatusCode != 200 {
		t.Errorf("Invalid status code: expecting %d, got %d", 200, response.StatusCode)
	}
	response.Body.Close()
}
