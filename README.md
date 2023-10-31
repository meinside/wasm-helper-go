# wasm-helper-go

[Wasm](https://webassembly.org/) helper library for Golang.

## Usage

```bash
$ go get -u github.com/meinside/wasm-helper-go
```

then

```go
// go:build js && wasm

package main

import (
	"fmt"
	"log"
	"syscall/js"

	wh "github.com/meinside/wasm-helper-go"
)

const (
	//debug = false
	debug = true
)

func main() {
	// get a new helper,
	helper := wh.New()
	helper.SetVerbose(debug) // set verbosity,

	// register functions for using in javascript
	helper.RegisterFunctions(map[string]wh.WasmFunction{
		"showAlert": func(this js.Value, args []js.Value) interface{} {
			if _, err := helper.Call("alert", args[0].String()); err != nil {
				log.Printf("failed to call function `alert`: %s", err)
			}

			return nil
		},
	})

	// alert window location,
	if location, err := helper.Get("window.location.href"); err == nil {
		if !location.IsUndefined() && !location.IsNull() {
			loc := location.String()

			if _, err := helper.Call("showAlert", fmt.Sprintf("window.location.href = %s", loc)); err != nil {
				log.Printf("failed to call function `showAlert`: %s", err)
			}
		}
	}

	// and wait...
	helper.Wait()
}
```

For more, see the example application and related files in [sample/](https://github.com/meinside/wasm-helper-go/tree/master/sample).

