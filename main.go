package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/cbodonnell/proxy-host/pkg/cache"
)

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxyCache *cache.Cache) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var proxy *httputil.ReverseProxy
		cached := proxyCache.Get(r.Host)
		if cached == nil {
			// TODO: check the database for the host and get the target url
			// if the host is not found in the database, return 404
			// if the host is found, create a new proxy with the target url
			// targetHost := "abcdefg.tunnel.farm"
			targetHost := "520cf64.dev.local:7880"
			url := &url.URL{
				Scheme: "http",
				Host:   targetHost,
			}
			newProxy := httputil.NewSingleHostReverseProxy(url)
			director := newProxy.Director
			newProxy.Director = func(r *http.Request) {
				director(r)
				r.Host = targetHost
				r.Header.Set("X-Proxy-Host", "true")
			}
			proxyCache.Set(r.Host, newProxy, 0)
			proxy = newProxy
		} else {
			// proxyCache.Extend(r.Host, 0) // wait until we can invalidate the cache
			proxy = cached.(*httputil.ReverseProxy)
		}

		proxy.ServeHTTP(w, r)
	}
}

func main() {
	proxyCache := cache.NewCache(5*time.Minute, 30*time.Second)
	defer proxyCache.StopCleanup()

	http.HandleFunc("/", ProxyRequestHandler(proxyCache))
	log.Fatal(http.ListenAndServe(":9999", nil))
}
