package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"strconv"

	"golang.org/x/net/proxy"
)

type handler func(map[string][]string) (string, error)

type Commander struct {
	ConstantPath string
}

func NewCommander(path string) *Commander {
	return &Commander{ ConstantPath: path }
}

var DEFAULT_HANDLER = func(request map[string][]string, c *Commander) (string, error) {
	call := strings.Join(request["call"], "")
	switch call {
		case "id":
			return c.GetHSLink(), nil
		case "send":
			rec := strings.Join(request["recepient"], "")
			cb := c.GetCallbackLink(rec)
			if cb == "" {
				return "transaction didn't happen", nil
			}
			msg := strings.Join(request["msg"], "")
			emsg := c.CipherMessage(rec, msg)
			tx, err := FormRawTxWithBlockchain(emsg, rec)
			if err != nil {
				fmt.Println(err)
				return "transaction didn't happen", err
			}
			link := c.GetHSLink()
			if saved := c.SaveMessage(msg, rec); saved != false {
				Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx)
			} else {
				return "", errors.New("Can't save message")
			}
			return tx, nil
		case "inbox":
			response := ""
			addr := strings.Join(request["address"], "")
			amount, err := strconv.Atoi(strings.Join(request["amount"], ""))
			if err != nil {
				return "[]", err
			}
			offset, err := strconv.Atoi(strings.Join(request["offset"], ""))
			if err != nil {
				return "[]", err
			}
			messages, err := c.GetMessages(addr, []int{amount, offset})
			if err != nil {
				return "[]", err
			}
			if len(messages) != 0 {
				for m := range messages {
					out, err := json.Marshal(messages[m])
					if err != nil {
						return "[]", err
					}
					response = fmt.Sprintf("%s%s,", response, string(out))
				}
				response = response[:len(response) - 1]
				response = "[" + response + "]"
			}
			return response, nil
		case "balanceOf":
			addr := strings.Join(request["address"], "")
			balance := GetBalance(addr)
			return balance, nil
		case "notify":
			cb := strings.Join(request["callback"], "")
			addr := c.GetAddressByLink(cb)
			tx := strings.Join(request["tx"], "")
			trimmedTx := strings.Split(tx, "x")[1]
			decodedTx, err := DecodeRawTx(trimmedTx)
			if err != nil {
				return "", err
			}
			res := c.DecipherMessage(addr, decodedTx)
			m := fmt.Sprintf("%s", res)
			if saved := c.SaveMessage(m, addr); saved != false {
				fmt.Printf(m)
				return "ok", nil
			}
			return "", errors.New("Can't save message")
		case "greeting":
			cb := strings.Join(request["callback"], "")
			cb = fmt.Sprintf("%s.onion", cb)
			existance := c.CheckExistance(cb)
			if existance != nil {
				return "already connected", nil
			}
			cipher := GenRandomString(32)
			hexedCipher := Hexify(cipher)
			link := c.GetHSLink()
			link = strings.Split(link, ".")[0]
			selfAddr := c.GetSelfAddress()
			formattedUrl := fmt.Sprintf("%s/?call=greetingOk", cb)
			formattedUrl = fmt.Sprintf("%s&callback=%s", formattedUrl, link)
			formattedUrl = fmt.Sprintf("%s&address=%s", formattedUrl, selfAddr)
			formattedUrl = fmt.Sprintf("%s&cipher=%s", formattedUrl, hexedCipher)
			response, err := Request(formattedUrl)
			if err != nil {
				fmt.Println(err)
			}
			err = c.WriteDownNewUser(cb, response, hexedCipher)
			if err != nil {
				return "can't save user", nil
			}
			return "ok", nil
		case "greetingOk":
			addr := strings.Join(request["address"], "")
			cb := strings.Join(request["callback"], "")
			cb = fmt.Sprintf("%s.onion", cb)
			cipher := strings.Join(request["cipher"], "")
			err := c.WriteDownNewUser(cb, addr, cipher)
			if err != nil {
				return "can't save user", nil
			}
			return c.GetSelfAddress(), nil
		default:
			return "unrecognized call", nil
		}
	}

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

func (c *Commander) GetHSLink() string {
	path := c.ConstantPath
	pathToHostname := path + "/hs/hostname"
	data, _ := ioutil.ReadFile(pathToHostname)
	link := strings.Split(string(data), "\n")[0]
	return link
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
		fmt.Println(err)
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println(string(b))
	return string(b), nil
}

func (c *Commander) RunTorAndHS() {
	switch runtime.GOOS {
	case "windows":
		return
	default:
		command := "cd "+c.ConstantPath+" && ./tor -f "+c.ConstantPath+"/torrc"
		exec.Command("sh", "-c", command).Output()
	}
}

func (c *Commander) RunRealServer() {
	server := &http.Server {
		Addr: ":4887",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response, err := DEFAULT_HANDLER(r.URL.Query(), c)
			if err != nil {
				response = "Error on the tor-side"
			}
			// sending back the response as web-server answer
			w.Write([]byte(response))
		})}
	server.ListenAndServe()
}
