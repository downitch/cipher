package main

import(
	"fmt"
	"strings"

	"./api"
)

func main() {

	httpsClient, err := api.RunGeth()
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		err = api.RunTorAndHS()
		if err != nil {
			fmt.Println(err)
		}
	}()

	link := api.GetHSLink()
	fmt.Println(link)

	api.RunRealServer(func(request map[string][]string) (string, error) {
		call := strings.Join(request["call"], "")
		switch call {
		case "id":
			response := "welcome!"
			return response, nil
		case "send":
			msg := strings.Join(request["msg"], "")
			rec := strings.Join(request["recepient"], "")
			key := strings.Join(request["key"], "")
			emsg, _ := api.EncodeMessage("/history/history", httpsClient, rec, msg)
			tx, err := api.SendMessageByBlockchain(httpsClient, key, emsg, rec)
			if err != nil {
				fmt.Println(err)
				return "transaction didn't happen", err
			}
			cb := api.GetCallbackLink(rec)
			go func() {
				api.Request(cb + "/?call=notify&callback=" + link + "&tx=" + tx)
			}()
			return tx, nil
		case "balanceOf":
			addr := strings.Join(request["address"], "")
			balance := api.GetBalance(httpsClient, addr)
			return balance, nil
		case "notify":
			// cb := strings.Join(request["callback"], "")
			tx := strings.Join(request["tx"], "")
			// emsg, err := api.ReadTx(tx)
			// dmsg := api.Decode(emsg, addr)
			return tx, nil
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
			callbackResponse, err := api.Request(cb + "/?call=greetingOk&callback=" + link + "&address=0x9De9223eb770E377ab148B8d37Fee348E8D691bC&cipher=" + hexedCipher)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(callbackResponse)
			return cipher, nil
		case "greetingOk":
			addr := strings.Join(request["address"], "")
			cb := strings.Join(request["callback"], "")
			cipher := strings.Join(request["cipher"], "")
			err := api.WriteDownNewUser(cb, addr, cipher)
			if err != nil {
				return "can't save user", nil
			}
			return cipher, nil
		default:
			return "unrecognized call", nil
		}
	})
}