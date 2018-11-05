package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ResponseJSON struct {
	Res string `json:"res"`
	Err string `json:"err"`
}

var DEFAULT_HANDLER = func(request map[string][]string, c *Commander) (string, error) {
	call := strings.Join(request["call"], "")
	switch call {
	case "id":
		id := c.GetHSLink()
		response := fmt.Sprintf(`{"res": "%s", "error": "nil"}`, id)
		return response, nil
	case "send":
		rec := strings.Join(request["recepient"], "")
		cb := c.GetCallbackLink(rec)
		if cb == "" {
			return `{"res": "nil", "error": "transaction didn't happen"}`, nil
		}
		msg := strings.Join(request["msg"], "")
		emsg := c.CipherMessage(rec, msg)
		tx, err := FormRawTxWithBlockchain(emsg, rec)
		if err != nil {
			return `{"res": "nil", "error": "can't form transaction"}`, nil
		}
		link := c.GetHSLink()
		if saved := c.SaveMessage(msg, rec); saved != false {
			Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx)
		} else {
			return `{"res": "nil", "error": "can't save message"}`, nil
		}
		return fmt.Sprintf(`{"res": "%s", "error": "nil"}`, tx), nil
	case "inbox":
		response := "nil"
		addr := strings.Join(request["address"], "")
		amount, err := strconv.Atoi(strings.Join(request["amount"], ""))
		if err != nil {
			return `{"res": "nil", "error": "can't convert amount to integer"}`, nil
		}
		offset, err := strconv.Atoi(strings.Join(request["offset"], ""))
		if err != nil {
			return `{"res": "nil", "error": "can't convert offset to integer"}`, nil
		}
		messages, err := c.GetMessages(addr, []int{amount, offset})
		if err != nil {
			return `{"res": "nil", "error": "can't get messages"}`, nil
		}
		if len(messages) != 0 {
			response = ""
			for m := range messages {
				out, err := json.Marshal(messages[m])
				if err != nil {
					response = fmt.Sprintf(`{"res": "nil", "error": "can't parse message #%d"}`, m)
					return response, nil
				}
				response = fmt.Sprintf("%s%s,", response, string(out))
			}
			response = response[:len(response) - 1]
			response = fmt.Sprintf("[%s]", response)
		}
		return fmt.Sprintf(`{"res": "%s", "error": "nil"}`, response), nil
	case "balanceOf":
		addr := strings.Join(request["address"], "")
		balance := GetBalance(addr)
		response := fmt.Sprintf(`{"res": "%s", "error": "nil"}`, balance)
		return response, nil
	case "notify":
		cb := strings.Join(request["callback"], "")
		addr := c.GetAddressByLink(cb)
		tx := strings.Join(request["tx"], "")
		trimmedTx := strings.Split(tx, "x")[1]
		decodedTx, err := DecodeRawTx(trimmedTx)
		if err != nil {
			return `{"res": "nil", "error": "can't decode tx"}`, err
		}
		res := c.DecipherMessage(addr, decodedTx)
		m := fmt.Sprintf("%s", res)
		if saved := c.SaveMessage(m, addr); saved != false {
			return `{"res": "ok", "error": "nil"}`, nil
		}
		return `{"res": "nil", "error": "can't save message"}`, errors.New("Can't save message")
	case "greeting":
		cb := strings.Join(request["callback"], "")
		cb = fmt.Sprintf("%s.onion", cb)
		if existance := c.CheckExistance(cb); existance != nil {
			return `{"res": "nil", "error": "already connected"}`, nil
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
			return fmt.Sprintf(`{"res": "nil", "error": "%s"}`, err), err
		}
		r := &ResponseJSON{}
		err = json.Unmarshal([]byte(response), r)
		if err != nil {
			return `{"res": "nil", "error": "can't parse response"}`, nil
		}
		err = c.WriteDownNewUser(cb, r.Res, hexedCipher)
		if err != nil {
			return `{"res": "nil", "error": "can't save user"}`, nil
		}
		return `{"res": "ok", "error": "nil"}`, nil
	case "greetingOk":
		addr := strings.Join(request["address"], "")
		cb := strings.Join(request["callback"], "")
		cb = fmt.Sprintf("%s.onion", cb)
		cipher := strings.Join(request["cipher"], "")
		err := c.WriteDownNewUser(cb, addr, cipher)
		if err != nil {
			return `{"res": "nil", "error": "can't save user"}`, nil
		}
		r := fmt.Sprintf(`{"res": "%s", "error": "nil"}`, c.GetSelfAddress())
		return r, nil
	default:
		return `{"res": "nil", "error": "unrecognized call"}`, nil
	}
}

func (c *Commander) RunRealServer() {
	server := &http.Server {
		Addr: ":4887",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response, _ := DEFAULT_HANDLER(r.URL.Query(), c)
			// sending back the response as web-server answer
			w.Write([]byte(response))
		})}
	server.ListenAndServe()
}