package proxy

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"syscall"

	"proxy/internal/api/usecase"

	requestUtils "proxy/pkg/http"
	"proxy/pkg/logger"
)

type Proxy struct {
	caCert  *x509.Certificate
	caKey   any
	Usecase usecase.Usecase
	Logger  logger.Logger
}

func NewProxy(caCertFile, caKeyFile string, u usecase.Usecase, l logger.Logger) (*Proxy, error) {
	caCert, caKey, err := loadX509KeyPair(caCertFile, caKeyFile)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		caCert:  caCert,
		caKey:   caKey,
		Usecase: u,
		Logger:  l,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleHTTPS(w, r)
	}

	p.handleHTTP(w, r)
}

type ProxyTransport struct {
	http.Transport
}

func (pt *ProxyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Del("Proxy-Connection")
	return pt.Transport.RoundTrip(r)
}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if bytes, err := httputil.DumpRequest(r, true); err == nil {
		p.Logger.Infof("incoming request:\n%s\n", string(bytes))
	}
	r.RequestURI = ""

	client := &http.Client{
		Transport: &ProxyTransport{},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	cpReq := *r

	resp, err := client.Do(r)
	if err != nil {
		p.Logger.Errorf("client error: %v", err)
		http.Error(w, "Failed to proxy %v", http.StatusBadRequest)
		return
	}

	if bytes, err := httputil.DumpResponse(resp, false); err == nil {
		p.Logger.Infof("target response:\n%s\n", string(bytes))
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		p.Logger.Fatal("http server doesn't support hijacking connection")
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		p.Logger.Fatal("http hijacking failed")
	}
	defer clientConn.Close()

	cpResp := *resp
	if err := resp.Write(clientConn); err != nil {
		p.Logger.Errorf("error writing response back: %v", err)
	}

	reqSave := requestUtils.ParseRequest(cpReq)
	respSave := requestUtils.ParseResponse(cpResp)

	id, err := p.Usecase.SaveRequest(r.Context(), *reqSave)
	if err != nil {
		p.Logger.Errorf("failed to insert request: %v", err)
	}

	respSave.RequestId = id

	if err := p.Usecase.SaveResponse(r.Context(), respSave); err != nil {
		p.Logger.Errorf("error while saving response: %v", err)
	}
}

func (p *Proxy) handleHTTPS(w http.ResponseWriter, proxyReq *http.Request) {
	p.Logger.Infof("CONNECT requested to %v (from %v)", proxyReq.Host, proxyReq.RemoteAddr)

	hj, ok := w.(http.Hijacker)
	if !ok {
		p.Logger.Fatal("http server doesn't support hijacking connection")
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		p.Logger.Fatal("http hijacking failed")
		return
	}
	defer clientConn.Close()

	host, _, err := net.SplitHostPort(proxyReq.Host)
	if err != nil {
		p.Logger.Fatal("error splitting host/port:", err)
		return
	}

	pemCert, pemKey := createCert([]string{host}, p.caCert, p.caKey, 240)
	tlsCert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		p.Logger.Fatal(err)
	}

	if _, err := clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		p.Logger.Fatal("error writing status to client:", err)
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{tlsCert},
	}

	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		p.handleTLSConnection(tlsConn, proxyReq)
	}()

	// Wait for the goroutine to finish before closing tlsConn
	wg.Wait()
}

func (p *Proxy) handleTLSConnection(tlsConn net.Conn, proxyReq *http.Request) {
	connReader := bufio.NewReader(tlsConn)

	defer func() {
		if r := recover(); r != nil {
			p.Logger.Errorf("Panic in handleTLSConnection: %v", r)
		}
	}()

	for {
		r, err := http.ReadRequest(connReader)
		if err == io.EOF {
			break
		} else if errors.Is(err, syscall.ECONNRESET) {
			p.Logger.Errorf("This is connection reset by peer error")
			break
		} else if err != nil {
			p.Logger.Fatal(proxyReq, err)
			break
		}

		go p.handleHTTPRequest(r, tlsConn, proxyReq)
	}
}

func (p *Proxy) handleHTTPRequest(r *http.Request, tlsConn net.Conn, proxyReq *http.Request) {
	if b, err := httputil.DumpRequest(r, false); err == nil {
		p.Logger.Infof("incoming request:\n%s\n", string(b))
	}

	cpReq := *r

	changeRequestToTarget(r, proxyReq.Host)

	client := http.Client{}

	resp, err := client.Do(r)
	if err != nil {
		p.Logger.Errorf("error sending request to target: %v", err)
		return
	}
	defer resp.Body.Close()

	cpResp := *resp
	if err := resp.Write(tlsConn); err != nil {
		p.Logger.Errorf("error writing response back: %v", err)
		return
	}

	reqSave := requestUtils.ParseRequest(cpReq)
	respSave := requestUtils.ParseResponse(cpResp)

	id, err := p.Usecase.SaveRequest(r.Context(), *reqSave)
	if err != nil {
		p.Logger.Errorf("failed to insert request: %v", err)
		return
	}

	respSave.RequestId = id

	if err := p.Usecase.SaveResponse(r.Context(), respSave); err != nil {
		p.Logger.Errorf("error while saving response: %v", err)
		return
	}
}
