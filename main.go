package main

import(
	"fmt"

	"./api"
)

func main() {
	go func() {
		err := api.RunTorAndHS()
		if err != nil {
			fmt.Println(err)
		}
	}()
	api.RunRealServer()
}