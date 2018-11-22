package main

import (
	//"flag"
	//"fmt"

	"gopkg.in/elazarl/goproxy.v1"
	//"io"
	"log"
	//"net"
	"net/http"

	"crypto/tls"
)

var handler *proxyHandler

func handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	resp := handler.Request(req, ctx)
	return req, resp
}

func handleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	handler.Response(resp, ctx)
	return resp
}

func main() {
	addr := ":8080"

	handler = NewHandler()

	proxy := goproxy.NewProxyHttpServer()

	cert, err := tls.LoadX509KeyPair("ca.pem", "key.pem")
	if err != nil {
		log.Fatalf("Unable to load certificate - %v", err)
	}

	proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return &goproxy.ConnectAction{
			Action:    goproxy.ConnectMitm,
			TLSConfig: goproxy.TLSConfigFromCA(&cert),
		}, host + ":443"
	})

	proxy.OnRequest().DoFunc(handleRequest)
	proxy.OnResponse().DoFunc(handleResponse)

	log.Fatal(http.ListenAndServe(addr, proxy))
}
