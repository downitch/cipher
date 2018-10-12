package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/cretz/bine/process/embedded"

	"golang.org/x/net/proxy"
)

type handler func(map[string][]string) (string, error)

func GetHSLink() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	pathToHostname := path + "/api/hs/hostname"
	data, err := ioutil.ReadFile(pathToHostname)
	link := strings.Split(string(data), "\n")[0]
	return link
}

func configureTorrc(path string) error {
  // formatting onion service setup
	settings := fmt.Sprintf("HiddenServiceDir %s/api/hs", path)
	settings = fmt.Sprintf("%s\nHiddenServicePort 80 127.0.0.1:4887", settings)
	// either creating a new file or writing to one that exists
	err := ioutil.WriteFile(path + "/api/torrc", []byte(settings), 0700)
	if err != nil {
		return err
	}
	// chmodding directory where application is running
	command := fmt.Sprintf("chmod 700 %s/api/hs", path)
	if _, err := exec.Command("sh", "-c", command).Output(); err != nil {
		return err
	}
	return nil
}

func Request(url string) (string, error) {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		return "", err
	}
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	req, err := http.NewRequest("GET", "http://" + url, nil)
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

func RunTorAndHS() error {
	var err error
	tor := embedded.NewCreator()
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	err = configureTorrc(path)
	if err != nil {
		return err
	}
	p, _ := tor.New(context.Background(), "-f", path + "/api/torrc", "--quiet")
	p.Start()
	return p.Wait()
}

func RunRealServer(hndlr handler) {
	server := &http.Server {
		Addr: ":4887",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response, err := hndlr(r.URL.Query())
			if err != nil {
				response = "Error on the tor-side"
			}
			// sending back the response as web-server answer
			w.Write([]byte(response))
		})}
	server.ListenAndServe()
}
