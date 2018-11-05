# Cipher API

![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)

Cipher Golang API is tightly used inside the project. It is responsible for all blockchain and TOR communications.

  - Getting Blockchain status
  - Contacts & Messages management
  - Cipher magic

# New Features!

  - Big files are now ready to be compressed before sending
  - File structure is optimized for other developers


You can also:
  - Communicate with blockchain, using Infura.io
  - Generate Private Keys used for signing transactions
  - Exchange passphrases using secure TOR Hidden Service protocol

# Example

This example shows how to setup correctly and start using Cipher via API and localhost.
Note, that all three ports (:9050, :9051, :4887) should be free and open.
API currently does not support any dynamic ports.

```go
package main

import(
	"os"

	"github.com/atomindustries/cphr_embedded"
)

func main() {
	path, _ := os.Getwd()
	commander := api.NewCommander(path)
	commander.ConfigureTorrc()
	go func() {
		commander.RunTorAndHS()
	}()
	commander.RunRealServer()
}
```

# Credits

Thanks to authors of SBC Encryption. May the free Internet be great again!