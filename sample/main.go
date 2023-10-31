// go:build js && wasm

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

	// show window.location.href,
	if location, err := helper.Get("window.location.href"); err == nil {
		if !location.IsUndefined() && !location.IsNull() {
			loc := location.String()

			if locationLabel, err := helper.Call("document.getElementById", "location"); err == nil {
				if !locationLabel.IsUndefined() && !locationLabel.IsNull() {
					if err := helper.SetOn(locationLabel, "innerHTML", loc); err != nil {
						log.Printf("failed to set innerHTML of locationLabel: %s", err)
					}
				}
			} else {
				log.Printf("failed to get location with id: %s", err)
			}
		}
	} else {
		log.Printf("failed to get window.location.href: %s", err)
	}

	// register functions,
	helper.RegisterFunctions(map[string]wh.WasmFunction{
		"initializeCounter": func(this js.Value, args []js.Value) any {
			// set initial counter value,
			if err := helper.Set("count", 0); err != nil {
				log.Printf("failed to set count: %s", err)
			}

			// and show it
			if count, err := helper.Get("count"); err == nil {
				if !count.IsUndefined() && !count.IsNull() {
					if countLabel, err := helper.Call("document.getElementById", "counter"); err == nil {
						if !countLabel.IsUndefined() && !countLabel.IsNull() {
							if err := helper.SetOn(countLabel, "innerHTML", count.Int()); err != nil {
								log.Printf("failed to set innerHTML of countLabel: %s", err)
							}
						}
					} else {
						log.Printf("failed to get counter by id: %s", err)
					}
				}
			} else {
				log.Printf("failed to get count: %s", err)
			}

			return nil
		},
		"increaseCounter": func(this js.Value, args []js.Value) any {
			// increase counter,
			if count, err := helper.Get("count"); err == nil {
				if !count.IsUndefined() && !count.IsNull() {
					count = js.ValueOf(count.Int() + 1)
					if err := helper.Set("count", count); err != nil { // count ++
						log.Printf("failed to set count: %s", err)
					}

					// and show it
					if countLabel, err := helper.Call("document.getElementById", "counter"); err == nil {
						if !countLabel.IsUndefined() && !countLabel.IsNull() {
							if err := helper.SetOn(countLabel, "innerHTML", count.Int()); err != nil {
								log.Printf("failed to set innerHTML of countLabel: %s", err)
							}
						}
					} else {
						log.Printf("failed to get counter by id: %s", err)
					}
				} else {
					log.Printf("failed to get count: %s", err)
				}
			}

			return nil
		},
	})

	// add event listeners,
	if button, err := helper.Call("document.getElementById", "button"); err == nil {
		if !button.IsUndefined() && !button.IsNull() {
			if _, err := helper.CallOn(button, "addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				log.Printf("button clicked")

				if _, err := helper.Call("increaseCounter"); err != nil {
					log.Printf("failed to call increaseCounter: %s", err)
				}

				return nil
			})); err != nil {
				log.Printf("failed to add click listener: %s", err)
			}
		}
	} else {
		log.Printf("failed to get button by id: %s", err)
	}

	// initialize,
	if _, err := helper.Call("initializeCounter"); err != nil {
		log.Printf("failed to call initializeCounter: %s", err)
	}

	// and wait...
	helper.Wait()
}
