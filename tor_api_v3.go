package api

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	// "runtime"
	"time"

	"golang.org/x/net/proxy"
)

// This function should be fired everytime Tor Hidden Service is running
func (c *Commander) ConfigureTorrc() error {
	path := c.ConstantPath
	// either creating a new file or writing to one that exists
	tcpPath := fmt.Sprintf("%s/tcp", path)
	hsPath := fmt.Sprintf("%s/hs", path)
	// formatting onion service setup
	settings := fmt.Sprintf("HiddenServiceDir %s", tcpPath)
	settings = fmt.Sprintf("%s\nHiddenServicePort 88 127.0.0.1:4888", settings)
	// formatting onion service setup
	settings = fmt.Sprintf("%s\n\nHiddenServiceDir %s", settings, hsPath)
	settings = fmt.Sprintf("%s\nHiddenServicePort 80 127.0.0.1:4887", settings)
	// either creating a new file or writing to one that exists
	err := ioutil.WriteFile(path + "/torrc", []byte(settings), 0644)
	if err != nil {
		return err
	}
	// chmodding directory where application is running
	correctPermission := int(0700)
	// switch runtime.GOOS {
	// case "windows":
	// 	correctPermission = int(0600)
	// default:
	// 	break
	// }
	if _, err := os.Stat(tcpPath); os.IsNotExist(err) {
   	os.Mkdir(hsPath, os.FileMode(correctPermission))
	}
	if _, err := os.Stat(hsPath); os.IsNotExist(err) {
   	os.Mkdir(hsPath, os.FileMode(correctPermission))
	}
	return nil
}

func Request(url string) (string, error) {
	// creating new dialer that will pass request over the proxy
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		return "", err
	}
	// creating all the structures, getting ready firing request
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	// requesting...
	req, err := http.NewRequest("GET", "http://" + url, nil)
	if err != nil {
		return "", err
	}
	// receiving response
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	// never forgetting to close response buffer at the end
	defer resp.Body.Close()
	// reading buffer into slice of bytes
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// parsed slice converted to string
	return string(b), nil
}

func RequestHTTPS(url string) (string, error) {
	// creating new dialer that will pass request over the proxy
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		return "", err
	}
	// creating all the structures, getting ready firing request
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	// requesting...
	req, err := http.NewRequest("GET", "https://" + url, nil)
	if err != nil {
		return "", err
	}
	// receiving response
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	// never forgetting to close response buffer at the end
	defer resp.Body.Close()
	// reading buffer into slice of bytes
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// parsed slice converted to string
	return string(b), nil
}

func RequestPostHTTPS(uri string, contentType string, bodyBuf *bytes.Buffer) (string, error) {
	// creating new dialer that will pass request over the proxy
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		return "", err
	}
	// creating all the structures, getting ready firing request
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	// requesting...
	req, err := http.NewRequest("POST", "https://" + uri, bodyBuf)
	req.Header.Set("Content-Type", contentType)
	if err != nil {
		return "", err
	}
	// req.Header.Add("Content-Type", contentType)
	// receiving response
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	// never forgetting to close response buffer at the end
	defer resp.Body.Close()
	// reading buffer into slice of bytes
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// parsed slice converted to string
	return string(b), nil
}

func RequestWithTimeout(url string) (string, error) {
	timeout := make(chan string, 1)
	// creating new dialer that will pass request over the proxy
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		return "", err
	}
	// creating all the structures, getting ready firing request
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	resp := &http.Response{}
	go func() {
		// requesting...
		req, err := http.NewRequest("GET", "http://" + url, nil)
		if err != nil {
			timeout <- "fail"
			return
		}
		// receiving response
		resp, err = httpClient.Do(req)
		if err != nil {
			timeout <- "fail"
			return
		}
		timeout <- "done"
	}()
	select {
	case res := <-timeout:
		if res != "fail" {
			// never forgetting to close response buffer at the end
			defer resp.Body.Close()
			// reading buffer into slice of bytes
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			// parsed slice converted to string
			return string(b), nil
		}
		return "", errors.New("No response for request")
	case <-time.After(15 * time.Second):
		return "", errors.New("Timeout reached")
	}
}

// This function is highly experimental, because windows is kinda weird
func (c *Commander) RunTorAndHS() {
	// switch runtime.GOOS {
	// case "windows":
	// 	return
	// default:
	command := "cd "+c.ConstantPath+" && ./tor --hush -f "+c.ConstantPath+"/torrc"
	out, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("%s\n", out)
	// }
}
