//go:build js && wasm

// Wasm helper library for Golang
//
// NOTE: open related files with GOOS and GOARCH environment variables like:
//    `$ GOOS=js GOARCH=wasm nvim __FILENAME__`

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
type WasmCallback func(this js.Value, args []js.Value) any

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
		prettified, _ := Prettify(callbacks)
		printLog("Registering callbacks: %s", prettified)
	}

	for name, callback := range callbacks {
		h.Set(name, js.FuncOf(callback))
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
	<-h.block

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
	if parent.IsUndefined() || parent.IsNull() {
		if h.verbose {
			prettified, _ := Prettify(names)
			printLog("Parent not given, using global for names: %s", prettified)
		}

		parent = js.Global()
	}

	// child
	child := parent.Get(names[0])
	if child.IsUndefined() {
		printLog("Error: '%s' is undefined", names[0])

		return child, nil
	} else if child.IsNull() {
		printLog("Error: '%s' is null", names[0])

		return child, nil
	}

	if h.verbose {
		prettified, _ := Prettify(names[1:])
		printLog("Recursing on child: %v with names: %s", child, prettified)
	}

	// recurse
	return h.get(child, names[1:])
}

// Set sets value for given name (eg: 'document.someparent.somechild.value')
func (h *WasmHelper) Set(name string, value any) bool {
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
		if parent.IsUndefined() || parent.IsNull() {
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
func (h *WasmHelper) SetOn(obj js.Value, propertyName string, value any) bool {
	if h.verbose {
		printLog("Setting value: %v on %v for name: '%s'", value, obj, propertyName)
	}

	// undefined / null check
	if obj.IsUndefined() || obj.IsNull() {
		printLog("Error: could not set value: '%v' for name: '%s' on object which is undefined or null", value, propertyName)

		return false
	}

	obj.Set(propertyName, value)

	return true
}

// Call calls a function with given name and arguments
func (h *WasmHelper) Call(name string, args ...any) js.Value {
	if h.verbose {
		prettified, _ := Prettify(args)
		printLog("Calling '%s' with arguments: %s", name, prettified)
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
	if parent.IsUndefined() || parent.IsNull() {
		printLog("Error: could not call: '%s' on a parent which is undefined or null", name)

		return parent
	}

	function := parent.Get(funcName)

	// undefined / null check
	if function.IsUndefined() || function.IsNull() {
		printLog("Error: could not call: '%s' which is undefined or null", funcName)

		return function
	}

	// type check
	if function.Type() != js.TypeFunction {
		printLog("Error: could not call '%s' which is not a function", name)

		return js.Undefined()
	}

	if h.verbose {
		prettified, _ := Prettify(args)
		printLog("Calling '%s' on %v with arguments: %s", funcName, parent, prettified)
	}

	return parent.Call(funcName, args...)
}

// CallOn calls a function on a object with given name and arguments
func (h *WasmHelper) CallOn(obj js.Value, funcName string, args ...any) js.Value {
	if h.verbose {
		prettified, _ := Prettify(args)
		printLog("Calling '%s' on %v with arguments: %s", funcName, obj, prettified)
	}

	if obj.IsUndefined() || obj.IsNull() {
		printLog("Error: could not call: '%s' on an object which is undefined or null", funcName)

		return obj
	}

	function := obj.Get(funcName)

	// undefined / null check
	if function.IsUndefined() || function.IsNull() {
		printLog("Error: could not call '%s' on %v which is undefined or null", funcName, obj)

		return function
	}

	// type check
	if function.Type() != js.TypeFunction {
		printLog("Error: could not call '%s' on %v which is not a function", funcName, obj)

		return js.Undefined()
	}

	if h.verbose {
		prettified, _ := Prettify(args)
		printLog("Calling '%s' on %v with arguments: %s", funcName, obj, prettified)
	}

	return obj.Call(funcName, args...)
}

// Invoke invokes given function with arguments
func (h *WasmHelper) Invoke(function js.Value, args ...any) js.Value {
	if h.verbose {
		prettified, _ := Prettify(args)
		printLog("Invoking %v with arguments: %s", function, prettified)
	}

	// undefined / null check
	if function.IsUndefined() || function.IsNull() {
		printLog("Error: could not invoke %v which is undefined or null", function)

		return function
	}

	// type check
	if function.Type() != js.TypeFunction {
		printLog("Error: could not invoke %v which is not a function", function)

		return js.Undefined()
	}

	if h.verbose {
		prettified, _ := Prettify(args)
		printLog("Invoking %v arguments: %s", function, prettified)
	}

	return function.Invoke(args...)
}

// print log to the console
func printLog(format string, v ...any) {
	log.Printf(format, v...)
}

// ToArray converts given value to an array.
func ToArray(value js.Value) ([]js.Value, error) {
	// undefined / null check
	if value.IsUndefined() || value.IsNull() {
		return nil, fmt.Errorf("cannot convert undefined or nil value to an array")
	}

	array := make([]js.Value, value.Length())
	for i := range array {
		array[i] = value.Index(i)
	}

	return array, nil
}

// Prettify returns a JSONized string of given value.
func Prettify(value any) (string, error) {
	var bytes []byte
	var err error
	if bytes, err = json.Marshal(value); err != nil {
		return fmt.Sprintf("%v", value), fmt.Errorf("failed to marshal given value: %s", err)
	}

	return string(bytes), nil
}
