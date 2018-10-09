package main

import(
	"fmt"
	"strings"

	"./api"
)

func main() {

	httpsClient, wsClient, err := api.RunGeth()
	if err != nil {
		fmt.Println("PIZDEC")
		fmt.Println(err)
	}

	go func(key string) {
		err := api.WatchBlockchain(wsClient, key)
		if err != nil {
			fmt.Println(err)
		}
	}("0x9De9223eb770E377ab148B8d37Fee348E8D691bC")

	err = api.Run(false, func(request map[string][]string) (string, error) {
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
			return tx, nil
		case "balanceOf":
			addr := strings.Join(request["address"], "")
			balance := api.GetBalance(httpsClient, addr)
			return balance, nil
		// case "notify":
			// addr := strings.Join(request["address"], "")
			// tx := strings.Join(request["tx"], "")
			// emsg, err := api.ReadTx(tx)
			// dmsg := api.Decode(emsg, addr)
			// dmsg := tx
			// return dmsg, nil
		// case "greeting":
		// 	addr := strings.Join(request["address"], "")

		default:
			return "unrecognized call", nil
		}
	})
	if err != nil {
		fmt.Println("PIZDEC2")
		fmt.Printf("%s", err)
	}
}