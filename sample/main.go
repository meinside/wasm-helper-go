// +build: js,wasm

package main

import (
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

	// show window.location,
	var windowLocation = "unknown"
	location := helper.Get("window.location")
	if location != js.Undefined() && location != js.Null() {
		windowLocation = location.String()

		locationLabel := helper.Call("document.getElementById", "location")
		if locationLabel != js.Undefined() && locationLabel != js.Null() {
			helper.SetOn(locationLabel, "innerHTML", windowLocation)
		}
	}

	// register callback functions,
	helper.RegisterCallbacks(map[string]wh.WasmCallback{
		"initializeCounter": func(args []js.Value) {
			// set initial counter value,
			helper.Set("count", 0)

			// and show it
			count := helper.Get("count")
			if count != js.Undefined() && count != js.Null() {
				countLabel := helper.Call("document.getElementById", "counter")
				if countLabel != js.Undefined() && countLabel != js.Null() {
					helper.SetOn(countLabel, "innerHTML", count.Int())
				}
			}
		},
		"increaseCounter": func(args []js.Value) {
			// increase counter,
			count := helper.Get("count")
			if count != js.Undefined() && count != js.Null() {
				count = js.ValueOf(count.Int() + 1)
				helper.Set("count", count) // count ++

				// and show it
				countLabel := helper.Call("document.getElementById", "counter")
				if countLabel != js.Undefined() && countLabel != js.Null() {
					helper.SetOn(countLabel, "innerHTML", count.Int())
				}
			}
		},
	})

	// add event listeners,
	button := helper.Call("document.getElementById", "button")
	if button != js.Undefined() && button != js.Null() {
		helper.CallOn(button, "addEventListener", "click", js.NewCallback(func(args []js.Value) {
			log.Printf("button clicked")

			helper.Call("increaseCounter")
		}))
	}

	// initialize,
	helper.Call("initializeCounter")

	// and wait...
	helper.Wait()
}
