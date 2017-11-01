package chrome

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

const listeningURL string = "127.0.0.1"

type forwardingProxy struct {
	targetURL *url.URL
	server    *httputil.ReverseProxy
	listener  net.Listener
	port      int
}

func (proxy *forwardingProxy) start() error {

	log.WithFields(log.Fields{"target-url": proxy.targetURL}).Debug("Initializing shitty forwarding proxy")

	// *Dont* verify remote certificates.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Start the proxy and assign our custom Transport
	proxy.targetURL.Path = "/" // set the path to / as this becomes the base path
	proxy.server = httputil.NewSingleHostReverseProxy(proxy.targetURL)
	proxy.server.Transport = transport

	// Get an open port for this proxy instance to run on.
	var err error
	proxy.listener, err = net.Listen("tcp", listeningURL+":0")
	if err != nil {
		return err
	}

	// Set the port we used so that the caller of this method
	// can discover where to find this proxy instance.
	proxy.port = proxy.listener.Addr().(*net.TCPAddr).Port
	log.WithFields(log.Fields{"target-url": proxy.targetURL, "listen-port": proxy.port}).
		Debug("forwarding proxy listening port")

	// Finally, the goroutine for the proxy service.
	go func() {

		log.WithFields(log.Fields{"target-url": proxy.targetURL, "listen-address": proxy.listener.Addr()}).
			Debug("Starting shitty forwarding proxy goroutine")

		// Create an isolated ServeMux
		//  ref: https://golang.org/pkg/net/http/#ServeMux
		httpServer := http.NewServeMux()
		httpServer.HandleFunc("/", proxy.handle)

		if err := http.Serve(proxy.listener, httpServer); err != nil {

			// Probably a better way to handle these cases. Meh.
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			// Looks like something is actually wrong
			log.WithFields(log.Fields{"err": err}).Error("Shitty forwarding proxy broke")
		}

	}()

	return nil
}

// handle gets called on each request. We use this to update the host header.
func (proxy *forwardingProxy) handle(w http.ResponseWriter, r *http.Request) {

	log.WithFields(log.Fields{"target-url": proxy.targetURL, "request": r.URL}).
		Debug("Making proxied request")

	// Replace the host so that the Host: header is correct
	r.Host = proxy.targetURL.Host

	proxy.server.ServeHTTP(w, r)
}

// Stops the proxy
func (proxy *forwardingProxy) stop() {

	log.WithFields(log.Fields{"target-url": proxy.targetURL, "port": proxy.port}).
		Debug("Stopping shitty forwarding proxy")

	proxy.listener.Close()
}
