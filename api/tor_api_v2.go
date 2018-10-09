package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"io/ioutil"

	"github.com/cretz/bine/process/embedded"
	"github.com/cretz/bine/tor"

	"golang.org/x/net/proxy"
)

type handler func(map[string][]string) (string, error)

type TorData struct {
	link string
}

func Request(url string) (string, error) {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		return "", err
	}
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	req, err := http.NewRequest("GET", "https://" + url, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Run(verbose bool, hndlr handler) error {
	var err error
	path, _ := os.Getwd()
	// Start tor
	startConf := &tor.StartConf{
		ProcessCreator: embedded.NewCreator(),
		TorrcFile: path + "/api/torrc",
		DataDir: path + "/api/hs",
		NoAutoSocksPort: true,
		ExtraArgs: []string{"--SocksPort", "9050"}}
	if verbose {
		startConf.DebugWriter = os.Stdout
	} else {
		startConf.ExtraArgs = append(startConf.ExtraArgs, "--quiet")
	}
	fmt.Println("Please wait a couple of minutes...")
	t, err := tor.Start(nil, startConf)
	if err != nil {
		return err
	}
	defer t.Close()
	// Wait at most a few minutes to publish the service
	listenCtx, listenCancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer listenCancel()
	// Create an onion service to listen on a random local port but show as
	// Do version 3, it's faster to set up
	onion, err := t.Listen(listenCtx, &tor.ListenConf{LocalPort: 4887, Detach: true, RemotePorts: []int{80}, Version3: true})
	fmt.Printf("%+v\n", onion)
	if err != nil {
		return err
	}
	defer onion.Close()
	// Start server asynchronously
	// fmt.Printf("Open Tor browser and navigate to http://%v.onion\n", onion.ID)
	fmt.Println("Press enter to exit")
	server := &http.Server {
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response, err := hndlr(r.URL.Query())
			if err != nil {
				response = "Error on the tor-side"
			}
			// sending back the response as web-server answer
			w.Write([]byte(response))
	})}
	defer server.Shutdown(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- server.Serve(onion) }()
	// Wait for key asynchronously
	go func() {
		fmt.Scanln()
		errCh <- nil
	}()
	// Stop when one happens
	defer fmt.Println("Closing")
	return <-errCh
}
