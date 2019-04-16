# wasm-helper-go

WASM helper library for Golang.

## Usage

```bash
$ go get -u github.com/meinside/wasm-helper-go
```

then

```go
// +build: js,wasm

package main

import (
	"fmt"
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

	// set callback functions
	helper.RegisterCallbacks(map[string]wh.WasmCallback{
		"showAlert": func(this js.Value, args []js.Value) interface{} {
			helper.Call("alert", args[0].String())

			return nil
		},
	})

	// alert window location,
	var windowLocation = "unknown"
	location := helper.Get("window.location")
	if location != js.Undefined() && location != js.Null() {
		windowLocation = location.String()

		helper.Call("showAlert", fmt.Sprintf("window.location = %s", windowLocation))
	}

	// and wait...
	helper.Wait()
}
```

For more, see the example application and related files in [sample/](https://github.com/meinside/wasm-helper-go/tree/master/sample).

