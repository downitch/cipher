package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"time"

	// "github.com/stackimpact/stackimpact-go"
)

type ResponseJSON struct {
	Res string `json:"res"`
	Err string `json:"err"`
}

var DEFAULT_ERROR = `{"res": "nil", "error": "can't convert struct to JSON"}`

var DEFAULT_HANDLER = func(request map[string][]string, c *Commander) (string, error) {
	call := strings.Join(request["call"], "")
	switch call {
	case "id":
		id := c.GetHSLink()
		return formResponse(id, ""), nil
	case "chats":
		chatsStr := ""
		var response []byte
		chats := c.GetChats()
		if len(chats) <= 0 {
			return formResponse("", "no chats found"), nil
		}
		for c := range chats {
			out, err := json.Marshal(chats[c])
			if err != nil {
				errResponse := fmt.Sprintf("can't parse chat #%d", c)
				return formResponse("", errResponse), nil
			}
			chatsStr = fmt.Sprintf("%s%s,", chatsStr, string(out))
		}
		chatsStr = chatsStr[:len(chatsStr) - 1]
		response = []byte(fmt.Sprintf(`{"res": [%s], "error": "nil"}`, chatsStr))
		return string(response), nil
	case "send":
		rec := strings.Join(request["recepient"], "")
		cb := c.GetLinkByAddress(rec)
		if cb == "" {
			return formResponse("", "transaction didn't happen"), nil
		}
		msg := strings.Join(request["msg"], "")
		emsg := c.CipherMessage(rec, msg)
		tx, err := FormRawTxWithBlockchain(emsg, rec)
		if err != nil {
			return formResponse("", "can't form transaction"), nil
		}
		link := c.GetHSLink()
		a := c.GetSelfAddress()
		id := c.SaveMessage(a, rec, "text", msg)
		if id == 0 {
			return formResponse("", "can't save message"), nil
		}
		go func() {
			r, err := Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx)
			if err != nil {
				c.UpdateFailedMessage(id, rec)
			}
			res := &ResponseJSON{}
			err = json.Unmarshal([]byte(r), res)
			if err != nil || res.Res != "ok" {
				c.UpdateFailedMessage(id, rec)
			}
		}()
		return formResponse(tx, ""), nil
	case "sysSend":
		rec := strings.Join(request["recepient"], "")
		cb := c.GetLinkByAddress(rec)
		if cb == "" {
			return formResponse("", "transaction didn't happen"), nil
		}
		msg := strings.Join(request["msg"], "")
		emsg := c.CipherMessage(rec, msg)
		tx, err := FormRawTxWithBlockchain(emsg, rec)
		if err != nil {
			return formResponse("", "can't form transaction"), nil
		}
		link := c.GetHSLink()
		a := c.GetSelfAddress()
		id := c.SaveMessage(a, rec, "system", msg)
		if id == 0 {
			return formResponse("", "can't save message"), nil
		}
		go func() {
			r, err := Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx + "&type=system")
			if err != nil {
				c.UpdateFailedMessage(id, rec)
			}
			res := &ResponseJSON{}
			err = json.Unmarshal([]byte(r), res)
			if err != nil || res.Res != "ok" {
				c.UpdateFailedMessage(id, rec)
			}
		}()
		return formResponse(tx, ""), nil
	case "fileSend":
		rec := strings.Join(request["recepient"], "")
		cb := c.GetLinkByAddress(rec)
		if cb == "" {
			return formResponse("", "transaction didn't happen"), nil
		}
		msg := strings.Join(request["msg"], "")
		emsg := c.CipherMessage(rec, msg)
		tx, err := FormRawTxWithBlockchain(emsg, rec)
		if err != nil {
			return formResponse("", "can't form transaction"), nil
		}
		link := c.GetHSLink()
		a := c.GetSelfAddress()
		id := c.SaveMessage(a, rec, "file", msg)
		if id == 0 {
			return formResponse("", "can't save message"), nil
		}
		go func() {
			r, err := Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx + "&type=file")
			if err != nil {
				c.UpdateFailedMessage(id, rec)
			}
			res := &ResponseJSON{}
			err = json.Unmarshal([]byte(r), res)
			if err != nil || res.Res != "ok" {
				c.UpdateFailedMessage(id, rec)
			}
		}()
		return formResponse(tx, ""), nil
	case "imageSend":
		rec := strings.Join(request["recepient"], "")
		cb := c.GetLinkByAddress(rec)
		if cb == "" {
			return formResponse("", "transaction didn't happen"), nil
		}
		msg := strings.Join(request["msg"], "")
		emsg := c.CipherMessage(rec, msg)
		tx, err := FormRawTxWithBlockchain(emsg, rec)
		if err != nil {
			return formResponse("", "can't form transaction"), nil
		}
		link := c.GetHSLink()
		a := c.GetSelfAddress()
		id := c.SaveMessage(a, rec, "file", msg)
		if id == 0 {
			return formResponse("", "can't save message"), nil
		}
		go func() {
			r, err := Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx + "&type=image")
			if err != nil {
				c.UpdateFailedMessage(id, rec)
			}
			res := &ResponseJSON{}
			err = json.Unmarshal([]byte(r), res)
			if err != nil || res.Res != "ok" {
				c.UpdateFailedMessage(id, rec)
			}
		}()
		return formResponse(tx, ""), nil
	case "resend":
		addr := strings.Join(request["address"], "")
		iid := strings.Join(request["id"], "")
		id, _ := strconv.Atoi(iid)
		msg := c.GetMessageById(addr, id)
		emsg := c.CipherMessage(addr, msg.Text)
		tx, err := FormRawTxWithBlockchain(emsg, addr)
		if err != nil {
			return formResponse("", "can't form transaction"), nil
		}
		cb := c.GetLinkByAddress(addr)
		if cb == "" {
			return formResponse("", "transaction didn't happen"), nil
		}
		link := c.GetHSLink()
		c.UpdateUnfailMessage(id, addr)
		r, err := RequestWithTimeout(cb + "/?call=notify&callback=" + link + "&tx=" + tx + "&type=" + msg.Type)
		if err != nil {
			c.UpdateFailedMessage(id, addr)
			return formResponse("", "user is offline"), nil
		}
		res := &ResponseJSON{}
		err = json.Unmarshal([]byte(r), res)
		if err != nil || res.Res != "ok" {
			c.UpdateFailedMessage(id, addr)
			return formResponse("", "recepient can't save message"), nil
		}
		return formResponse(tx, ""), nil
	case "inbox":
		response := "nil"
		addr := strings.Join(request["address"], "")
		messages, err := c.GetChatHistory(addr)
		if err != nil {
			return `{"res": [], "error": "can't get messages"}`, nil
		}
		if len(messages) != 0 {
			response = ""
			for m := range messages {
				out, err := json.Marshal(messages[m])
				if err != nil {
					response = fmt.Sprintf(`{"res": [],
						"error": "can't parse message #%d"}`, m)
					return response, nil
				}
				response = fmt.Sprintf("%s%s,", response, string(out))
			}
			response = response[:len(response) - 1]
		}
		if response == "nil" {
			return `{"res": [], "error": "nil"}`, nil
		}
		go func() {
			l := c.GetLinkByAddress(addr)
			a := c.GetSelfAddress()
			RequestWithTimeout(fmt.Sprintf("%s/?call=inboxFired&address=%s", l, a))
		}()
		return fmt.Sprintf(`{"res": [%s], "error": "nil"}`, response), nil
	case "inboxFired":
		addr := strings.Join(request["address"], "")
		c.UpdateSentMessages(addr)
		return formResponse("ok", ""), nil
	case "balanceOf":
		addr := strings.Join(request["address"], "")
		balance := GetBalance(addr)
		return formResponse(balance, ""), nil
	case "notify":
		cb := strings.Join(request["callback"], "")
		t := strings.Join(request["type"], "")
		if t == "" {
			t = "text"
		}
		addr := c.GetAddressByLink(cb)
		if addr == "" {
			return formResponse("", "no such user found"), nil
		}
		tx := strings.Join(request["tx"], "")
		trimmedTx := strings.Split(tx, "x")[1]
		decodedTx, err := DecodeRawTx(trimmedTx)
		if err != nil {
			return formResponse("", "can't decode tx"), nil
		}
		res := c.DecipherMessage(addr, decodedTx)
		m := fmt.Sprintf("%s", res)
		saved := c.SaveMessage(addr, addr, t, m)
		if saved > 0 {
			go func(c *Commander, addr string) {
				time.Sleep(time.Second * 5)
				c.UpdatedSelfNewMessages(addr)
			}(c, addr)
			return formResponse("ok", ""), nil
		}
		return formResponse("", "can't save message"), nil
	case "greeting":
		cb := strings.Join(request["callback"], "")
		cb = fmt.Sprintf("%s.onion", cb)
		if existance := c.CheckExistance(cb); existance != false {
			return formResponse("", "already connected"), nil
		}
		cipher := GenRandomString(32)
		hexedCipher := Hexify(cipher)
		link := strings.Split(c.GetHSLink(), ".")[0]
		selfAddr := c.GetSelfAddress()
		formattedUrl := fmt.Sprintf(`%s/?call=greetingOk&callback=%s&address=%s&cipher=%s`, cb, link, selfAddr, hexedCipher)
		fmt.Println(formattedUrl)
		response, err := Request(formattedUrl)
		if err != nil {
			return formResponse("", err.Error()), nil
		}
		r := &ResponseJSON{}
		if err = json.Unmarshal([]byte(response), r); err != nil {
			return formResponse("", "can't parse response"), nil
		}
		if err = c.AddNewUser(&NewUser{cb, r.Res, hexedCipher}); err != nil {
			return formResponse("", "can't save user"), nil
		}
		return formResponse("ok", ""), nil
	case "greetingOk":
		addr := strings.Join(request["address"], "")
		cb := fmt.Sprintf("%s.onion", strings.Join(request["callback"], ""))
		cipher := strings.Join(request["cipher"], "")
		if err := c.AddNewUser(&NewUser{cb, addr, cipher}); err != nil {
			return formResponse("", "can't save user"), nil
		}
		return formResponse(c.GetSelfAddress(), ""), nil
	default:
		return formResponse("", "unrecognized call"), nil
	}
}

func formResponse(response string, err string) string {
	if err != "" {
		responseStruct := ResponseJSON{"nil", err}
		responseStructStringified, e := json.Marshal(responseStruct)
		if e != nil {
			responseStruct = ResponseJSON{"nil", err}
			responseStructStringified, e = json.Marshal(responseStruct)
			if e != nil {
				responseStructStringified = []byte(DEFAULT_ERROR)
			}
		}
		return string(responseStructStringified)
	}
	responseStruct := ResponseJSON{response, "nil"}
	responseStructStringified, e := json.Marshal(responseStruct)
	if e != nil {
		responseStruct = ResponseJSON{"nil", err}
		responseStructStringified, e = json.Marshal(responseStruct)
		if e != nil {
			responseStructStringified = []byte(DEFAULT_ERROR)
		}
	}
	return string(responseStructStringified)
}

func (c *Commander) RunRealServer() {
	// stackimpact.Start(stackimpact.Options{
	//   AgentKey: "058d0f9bfc13ffc3ab893174159224c87f8c0c4e",
	//   AppName: "MyGoApp"})

	// Here we define our HTTP web-server that will be visible from darkweb
	server := &http.Server {
		Addr: ":4887",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response, _ := DEFAULT_HANDLER(r.URL.Query(), c)
			// sending back the response as web-server answer
			w.Write([]byte(response))
		})}
	server.ListenAndServe()
}