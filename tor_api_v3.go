package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"golang.org/x/net/proxy"
)

// This function should be fired everytime Tor Hidden Service is running
func (c *Commander) ConfigureTorrc() error {
	path := c.ConstantPath
	hsPath := fmt.Sprintf("%s/hs", path)
  // formatting onion service setup
	settings := fmt.Sprintf("HiddenServiceDir %s", hsPath)
	settings = fmt.Sprintf("%s\nHiddenServicePort 80 127.0.0.1:4887", settings)
	// either creating a new file or writing to one that exists
	err := ioutil.WriteFile(path + "/torrc", []byte(settings), 0644)
	if err != nil {
		return err
	}
	// chmodding directory where application is running
	switch runtime.GOOS {
	case "windows":
		return os.Chmod(hsPath, 0600)
	default:
		return os.Chmod(hsPath, 0700)
	}
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

// This function is highly experimental, because windows is kinda weird
func (c *Commander) RunTorAndHS() {
	switch runtime.GOOS {
	case "windows":
		return
	default:
		command := "cd "+c.ConstantPath+" && ./tor --hush -f "+c.ConstantPath+"/torrc"
		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			fmt.Printf(err.Error())
		}
		fmt.Printf("%s\n", out)
	}
}
