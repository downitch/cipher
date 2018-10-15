package main

import(
	"fmt"
	"strings"

	"./api"
)

func main() {

	go func() {
		err := api.RunTorAndHS()
		if err != nil {
			fmt.Println(err)
		}
	}()
	
	api.RunRealServer(func(request map[string][]string) (string, error) {
		call := strings.Join(request["call"], "")
		switch call {
		case "id":
			return api.GetHSLink(), nil
		case "send":
			rec := strings.Join(request["recepient"], "")
			cb := api.GetCallbackLink(rec)
			if cb == "" {
				return "transaction didn't happen", nil
			}
			msg := strings.Join(request["msg"], "")
			emsg, _ := api.CipherMessage(rec, msg)
			tx, err := api.FormRawTxWithBlockchain(emsg, rec)
			if err != nil {
				fmt.Println(err)
				return "transaction didn't happen", err
			}
			link := api.GetHSLink()
			api.Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx)
			return tx, nil
		case "balanceOf":
			addr := strings.Join(request["address"], "")
			balance := api.GetBalance(addr)
			return balance, nil
		case "notify":
			cb := strings.Join(request["callback"], "")
			tx := strings.Join(request["tx"], "")
			trimmedTx := strings.Split(tx, "x")[1]
			decodedTx, err := api.DecodeRawTx(trimmedTx)
			if err != nil {
				return "", err
			}
			fmt.Println(cb)
			fmt.Println(decodedTx)
			// emsg, err := api.ReadTx(tx)
			// dmsg := api.Decode(emsg, addr)
			return "ok", nil
		case "greeting":
			addr := strings.Join(request["address"], "")
			cb := strings.Join(request["callback"], "")
			existance := api.CheckExistance(addr)
			if existance != nil {
				return "already connected", nil
			}
			cipher := api.GenRandomString(32)
			hexedCipher := api.Hexify(cipher)
			err := api.WriteDownNewUser(cb, addr, hexedCipher)
			if err != nil {
				return "can't save user", nil
			}
			link := api.GetHSLink()
			selfAddr := api.GetSelfAddress()
			formattedUrl := fmt.Sprintf("%s/?call=greetingOk", cb)
			formattedUrl = fmt.Sprintf("%s&callback=%s", formattedUrl, link)
			formattedUrl = fmt.Sprintf("%s&address=%s", formattedUrl, selfAddr)
			formattedUrl = fmt.Sprintf("%s&cipher=%s", formattedUrl, hexedCipher)
			response, err := api.Request(formattedUrl)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(response)
			return cipher, nil
		case "greetingOk":
			addr := strings.Join(request["address"], "")
			cb := strings.Join(request["callback"], "")
			cipher := strings.Join(request["cipher"], "")
			err := api.WriteDownNewUser(cb, addr, cipher)
			if err != nil {
				return "can't save user", nil
			}
			return "ok", nil
		default:
			return "unrecognized call", nil
		}
	})
}