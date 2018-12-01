package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"time"
)

type ResponseJSON struct {
	Res string `json:"res"`
	Err string `json:"err"`
}

const DEFAULT_ERROR = `{"res": "nil", "error": "can't convert struct to JSON"}`

var DEFAULT_HANDLER = func(request map[string][]string, c *Commander) string {
	switch strings.Join(request["call"], "") {
	case "id":
		return formResponse(c.GetHSLink(), "")
	case "chats":
		return c.FormChats()
	case "send":
		return c.SendMessageWithType(request, "text")
	case "sysSend":
		return c.SendMessageWithType(request, "system")
	case "fileSend":
		return c.SendMessageWithType(request, "file")
	case "imageSend":
		return c.SendMessageWithType(request, "image")
	case "audioSend":
		return c.SendMessageWithType(request, "audio")
	case "resend":
		return c.SendMessageWithType(request, "")
	case "inbox":
		return c.ProcessInbox(request)
	case "inboxFired":
		address := strings.Join(request["address"], "")
		c.UpdateSentMessages(address)
		return formResponse("ok", "")
	case "notify":
		return c.ProcessNotification(request)
	case "greeting":
		return c.InitiateGreeting(request)
	case "greetingOk":
		return c.ProcessGreeting(request)
	default:
		return formResponse("", "unrecognized call")
	}
}

func (c *Commander) RunRealServer() {
	// Here we define our HTTP web-server that will be visible from darkweb
	server := &http.Server {
		Addr: ":4887",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(DEFAULT_HANDLER(r.URL.Query(), c)))
		}),
	}
	server.ListenAndServe()
}

func (c *Commander) FormChats() string {
	chatsStr := ""
	chats := c.GetChats()
	if len(chats) <= 0 {
		return formResponse("", "no chats found")
	}
	for i := range chats {
		out, _ := json.Marshal(chats[i])
		chatsStr = fmt.Sprintf("%s%s,", chatsStr, string(out))
	}
	chatsStr = chatsStr[:len(chatsStr) - 1]
	return fmt.Sprintf(`{"res": [%s], "error": "nil"}`, chatsStr)
}

func (c *Commander) SendMessageWithType(request map[string][]string, t string) string {
	var msg string
	var id int
	rec := strings.Join(request["recepient"], "")
	mid := strings.Join(request["id"], "")
	cb := c.GetLinkByAddress(rec)
	link := c.GetHSLink()
	if cb == "" {
		return formResponse("", "transaction didn't happen")
	}
	if mid == "" {
		msg = strings.Join(request["msg"], "")
		a := c.GetSelfAddress()
		id = c.SaveMessage(a, rec, t, msg)
		if id == 0 {
			return formResponse("", "can't save message")
		}
	} else {
		id, _ = strconv.Atoi(mid)
		addr := strings.Join(request["address"], "")
		message := c.GetMessageById(addr, id)
		msg = message.Text
		t = message.Type
	}
	emsg := c.CipherMessage(rec, msg)
	tx, err := FormRawTxWithBlockchain(emsg, rec)
	if err != nil {
		return formResponse("", "can't form transaction")
	}
	go func() {
		uri := fmt.Sprintf("%s/?call=notify&callback=%s&tx=%s&type=%s", cb, link, tx, t)
		r, err := Request(uri)
		if err != nil {
			c.UpdateFailedMessage(id, rec)
		}
		res := &ResponseJSON{}
		err = json.Unmarshal([]byte(r), res)
		if err != nil || res.Res != "ok" {
			c.UpdateFailedMessage(id, rec)
		}
	}()
	return formResponse(tx, "")
}

func (c *Commander) ProcessInbox(request map[string][]string) string {
	var msgs []NewMessage
	var response string
	var out []byte
	var err error
	address := strings.Join(request["address"], "")
	if msgs, err = c.GetChatHistory(address); err != nil || len(msgs) == 0 {
		return `{"res": [], "error": "can't get messages"}`
	}
	for i := range msgs {
		if out, err = json.Marshal(msgs[i]); err != nil {
			return `{"res": [], "error": "can't parse message"}`
		}
		response = fmt.Sprintf("%s%s,", response, string(out))
	}
	response = response[:len(response) - 1]
	go func() {
		l := c.GetLinkByAddress(address)
		a := c.GetSelfAddress()
		uri := fmt.Sprintf("%s/?call=inboxFired&address=%s", l, a)
		Request(uri)
	}()
	return fmt.Sprintf(`{"res": [%s], "error": "nil"}`, response)
}

func (c *Commander) ProcessNotification(request map[string][]string) string {
	var t string
	var address string
	cb := strings.Join(request["callback"], "")
	if t = strings.Join(request["type"], ""); t == "" {
		t = "text"
	}
	if address = c.GetAddressByLink(cb); address == "" {
		return formResponse("", "no such user found")
	}
	tx := strings.Join(request["tx"], "")
	trimmedTx := strings.Split(tx, "x")[1]
	decodedTx, err := DecodeRawTx(trimmedTx)
	if err != nil {
		return formResponse("", "can't decode tx")
	}
	m := string(c.DecipherMessage(address, decodedTx))
	if saved := c.SaveMessage(address, address, t, m); saved == 0 {
		return formResponse("", "can't save message")
	}
	go func(c *Commander, address string) {
		time.Sleep(time.Second * 5)
		c.UpdatedSelfNewMessages(address)
	}(c, address)
	return formResponse("ok", "")
}

func (c *Commander) InitiateGreeting(request map[string][]string) string {
	callback := strings.Join(request["callback"], "")
	cb := fmt.Sprintf("%s.onion", callback)
	if existance := c.CheckExistance(cb); existance != false {
		return formResponse("", "already connected")
	}
	cipher := GenRandomString(32)
	hexedCipher := Hexify(cipher)
	link := strings.Split(c.GetHSLink(), ".")[0]
	selfAddr := c.GetSelfAddress()
	formattedUrl := fmt.Sprintf(`%s/?call=greetingOk&callback=
	%s&address=%s&cipher=%s`, cb, link, selfAddr, hexedCipher)
	response, err := Request(formattedUrl)
	if err != nil {
		return formResponse("", err.Error())
	}
	r := &ResponseJSON{}
	if err = json.Unmarshal([]byte(response), r); err != nil {
		return formResponse("", "can't parse response")
	}
	if err = c.AddNewUser(&NewUser{cb, r.Res, hexedCipher}); err != nil {
		return formResponse("", "can't save user")
	}
	return formResponse("ok", "")
}

func (c *Commander) ProcessGreeting(request map[string][]string) string {
	addr     := strings.Join(request["address"], "")
	callback := strings.Join(request["callback"], "")
	cipher   := strings.Join(request["cipher"], "")
	cb       := fmt.Sprintf("%s.onion", callback)
	newUser  := &NewUser{
		cb,
		addr,
		cipher,
	}
	if err := c.AddNewUser(newUser); err != nil {
		return formResponse("", "can't save user")
	}
	return formResponse(c.GetSelfAddress(), "")
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