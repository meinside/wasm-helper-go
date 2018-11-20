// WASM helper library for Golang
//
// (open files with: `$ GOOS=js GOARCH=wasm vi __FILENAME__`)

// +build js,wasm

package wasmhelper

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"syscall/js"
)

// WasmHelper struct
type WasmHelper struct {
	block   chan struct{}
	verbose bool
}

// WasmCallback function type
type WasmCallback func(args []js.Value)

// New returns a new WasmHelper struct
func New() *WasmHelper {
	return &WasmHelper{
		block:   make(chan struct{}, 1),
		verbose: false,
	}
}

// SetVerbose sets verbosity of this helper
func (h *WasmHelper) SetVerbose(isVerbose bool) {
	h.verbose = isVerbose
}

// RegisterCallbacks registers given callback functions
func (h *WasmHelper) RegisterCallbacks(callbacks map[string]WasmCallback) {
	if h.verbose {
		printLog("Registering callbacks: %s", Prettify(callbacks))
	}

	for name, callback := range callbacks {
		h.Set(name, js.NewCallback(callback))
	}
}

// Wait blocks until stopped manually, for long-running routines
//
// May panic when there is no registered callback or event listener.
func (h *WasmHelper) Wait() {
	if h.verbose {
		printLog("Waiting...")
	}

	// wait...
	select {
	case <-h.block:
		break
	}

	if h.verbose {
		printLog("Stopped waiting")
	}
}

// Stop stops blocking
func (h *WasmHelper) Stop() {
	if h.verbose {
		printLog("Stopping waiting...")
	}

	h.block <- struct{}{}
}

// Get retrieves value for given name (eg: 'document.someparent.somechild.value')
func (h *WasmHelper) Get(name string) js.Value {
	if h.verbose {
		printLog("Getting value for name: '%s'", name)
	}

	names := strings.Split(name, ".")
	count := len(names)

	if count > 0 {
		value, _ := h.get(js.Null(), names)

		if h.verbose {
			printLog("Got value: %v for name: '%s'", value, name)
		}

		return value
	}

	printLog("Error: could not get value, given name is empty")

	return js.Undefined()
}

// get value from names recursively
func (h *WasmHelper) get(parent js.Value, names []string) (value js.Value, remainingNames []string) {
	if len(names) == 0 {
		return parent, nil
	}

	// parent
	if parent == js.Undefined() || parent == js.Null() {
		if h.verbose {
			printLog("Parent not given, using global for names: %s", Prettify(names))
		}

		parent = js.Global()
	}

	// child
	child := parent.Get(names[0])
	if child == js.Undefined() {
		printLog("Error: '%s' is undefined", names[0])

		return child, nil
	} else if child == js.Null() {
		printLog("Error: '%s' is null", names[0])

		return child, nil
	}

	if h.verbose {
		printLog("Recursing on child: %v with names: %s", child, Prettify(names[1:]))
	}

	// recurse
	return h.get(child, names[1:])
}

// Set sets value for given name (eg: 'document.someparent.somechild.value')
func (h *WasmHelper) Set(name string, value interface{}) bool {
	if h.verbose {
		printLog("Setting value: %v for name: '%s'", value, name)
	}

	names := strings.Split(name, ".")
	count := len(names)

	var lastName string
	var parent js.Value

	if count >= 2 {
		parentNames := names[:count-1]
		parent, _ = h.get(js.Null(), parentNames)

		// undefined / null check
		if parent == js.Undefined() || parent == js.Null() {
			printLog("Error: could not set value, '%s' is undefined or null", strings.Join(parentNames, "."))

			return false
		}

		lastName = names[count-1]
	} else if count == 1 {
		parent = js.Global()
		lastName = names[0]
	} else {
		printLog("Error: could not set value, given name is empty")

		return false
	}

	// set value
	parent.Set(lastName, value)

	return true
}

// SetOn sets value for given property name on given object
func (h *WasmHelper) SetOn(obj js.Value, propertyName string, value interface{}) bool {
	if h.verbose {
		printLog("Setting value: %v on %v for name: '%s'", value, obj, propertyName)
	}

	// undefined / null check
	if obj == js.Undefined() || obj == js.Null() {
		printLog("Error: could not set value: '%v' for name: '%s' on object which is undefined or null", value, propertyName)

		return false
	}

	obj.Set(propertyName, value)

	return true
}

// Call calls a function with given name and arguments
func (h *WasmHelper) Call(name string, args ...interface{}) js.Value {
	if h.verbose {
		printLog("Calling '%s' with arguments: %s", name, Prettify(args))
	}

	names := strings.Split(name, ".")
	parentNames := names[:len(names)-1]
	funcName := names[len(names)-1]

	var parent js.Value
	if len(names) >= 2 {
		parent = h.Get(strings.Join(parentNames, "."))
	} else {
		parent = js.Global()
	}

	// undefined / null check
	if parent == js.Undefined() || parent == js.Null() {
		printLog("Error: could not call: '%s' on a parent which is undefined or null", name)

		return parent
	}

	function := parent.Get(funcName)

	// undefined / null check
	if function == js.Undefined() || function == js.Null() {
		printLog("Error: could not call: '%s' which is undefined or null", funcName)

		return function
	}

	// type check
	if function.Type() != js.TypeFunction {
		printLog("Error: could not call '%s' which is not a function", name)

		return js.Undefined()
	}

	if h.verbose {
		printLog("Calling '%s' on %v with arguments: %s", funcName, parent, Prettify(args))
	}

	return parent.Call(funcName, args...)
}

// CallOn calls a function on a object with given name and arguments
func (h *WasmHelper) CallOn(obj js.Value, funcName string, args ...interface{}) js.Value {
	if h.verbose {
		printLog("Calling '%s' on %v with arguments: %s", funcName, obj, Prettify(args))
	}

	if obj == js.Undefined() || obj == js.Null() {
		printLog("Error: could not call: '%s' on an object which is undefined or null", funcName)

		return obj
	}

	function := obj.Get(funcName)

	// undefined / null check
	if function == js.Undefined() || function == js.Null() {
		printLog("Error: could not call '%s' on %v which is undefined or null", funcName, obj)

		return function
	}

	// type check
	if function.Type() != js.TypeFunction {
		printLog("Error: could not call '%s' on %v which is not a function", funcName, obj)

		return js.Undefined()
	}

	if h.verbose {
		printLog("Calling '%s' on %v with arguments: %s", funcName, obj, Prettify(args))
	}

	return obj.Call(funcName, args...)
}

// Invoke invokes given function with arguments
func (h *WasmHelper) Invoke(function js.Value, args ...interface{}) js.Value {
	if h.verbose {
		printLog("Invoking %v with arguments: %s", function, Prettify(args))
	}

	// undefined / null check
	if function == js.Undefined() || function == js.Null() {
		printLog("Error: could not invoke %v which is undefined or null", function)

		return function
	}

	// type check
	if function.Type() != js.TypeFunction {
		printLog("Error: could not invoke %v which is not a function", function)

		return js.Undefined()
	}

	if h.verbose {
		printLog("Invoking %v arguments: %s", function, Prettify(args))
	}

	return function.Invoke(args...)
}

// print log to the console
func printLog(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// ToArray converts given value to array (returns nil on error)
func ToArray(value js.Value) []js.Value {
	// undefined / null check
	if value == js.Undefined() || value == js.Null() {
		printLog("Error: could not convert undefined or nil value to array")

		return nil
	}

	array := make([]js.Value, value.Length())
	for i := range array {
		array[i] = value.Index(i)
	}

	return array
}

// Prettify returns JSONized string of given value
func Prettify(value interface{}) string {
	var bytes []byte
	var err error
	if bytes, err = json.Marshal(value); err != nil {
		printLog("Failed to marshal given value: %+v", value)

		return fmt.Sprintf("%v", value)
	}

	return string(bytes)
}
