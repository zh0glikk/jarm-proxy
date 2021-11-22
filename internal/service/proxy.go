package service

import (
	"errors"
	"fmt"
	"github.com/google/jsonapi"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/jarm-proxy/internal/config"
	"github.com/jarm-proxy/internal/jarm"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
	log        *logan.Entry
	jarmConfig *config.JarmConfig
}

func newProxy(log *logan.Entry, jarmConfig *config.JarmConfig) *proxy {
	return &proxy{log: log, jarmConfig: jarmConfig}
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	ok, details, err := p.isAllowed(req.URL.String())
	if err != nil {
		p.log.WithError(err).Error("seems to be bad target")
		ape.Render(wr, problems.InternalError())
		return
	}
	if !ok {
		p.log.Info("not allowed to connect")
		ape.Render(wr, &jsonapi.ErrorObject{
			Title:  http.StatusText(http.StatusForbidden),
			Status: fmt.Sprintf("%d", http.StatusForbidden),
			Detail: details,
		})
		return
	}

	client := &http.Client{}

	req.RequestURI = ""

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(req)
	if err != nil {
		p.log.WithError(err).Error("failed to create request")
		ape.Render(wr, problems.InternalError())
		return
	}
	defer resp.Body.Close()
	p.log.Info(resp.Status)
	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func (p *proxy) isAllowed(address string) (bool, string, error) {
	address = strings.Trim(address, "http://")
	address = strings.Trim(address, "https://")

	address, port := parseAddress(address)
	if port == nil {
		port = &p.jarmConfig.DefaultPort
	}

	result := jarm.Fingerprint(jarm.Target{
		Host: address,
		Port: *port,
	})
	if result == nil {
		return false, "", errors.New("failed to create fingerptint")
	}

	p.log.WithFields(logan.F{
		"host": result.Target.Host,
		"hash": result.Hash,
	}).Info("got result")

	fingerPrints := p.jarmConfig.Fingerprints

	for _, element := range fingerPrints {
		if result.Hash == element {
			p.log.Debug("should not be connected")
			return false, result.Hash, nil
		}
	}
	return true, "", nil
}

func parseAddress(address string) (string, *int) {
	res := strings.Split(address, ":")
	if len(res) == 2 {
		port := cast.ToInt(res[1])
		return res[0], &port
	}

	return address, nil
}
