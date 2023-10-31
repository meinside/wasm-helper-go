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
	"sync"
	"syscall/js"
)

// WasmHelper struct
type WasmHelper struct {
	block chan struct{}
	funcs map[string]js.Func
	lock  sync.Mutex

	verbose bool
}

// WasmFunction function type
type WasmFunction func(this js.Value, args []js.Value) any

// New returns a new WasmHelper struct.
func New() *WasmHelper {
	return &WasmHelper{
		block:   make(chan struct{}, 1),
		funcs:   make(map[string]js.Func),
		verbose: false,
	}
}

// SetVerbose sets verbosity of this helper.
func (h *WasmHelper) SetVerbose(isVerbose bool) {
	h.verbose = isVerbose
}

// RegisterFunctions registers given functions.
func (h *WasmHelper) RegisterFunctions(fns map[string]WasmFunction) {
	if h.verbose {
		prettified, _ := Prettify(fns)

		l("Registering functions: %s", prettified)
	}

	for name, fn := range fns {
		f := js.FuncOf(fn)

		h.lock.Lock()

		if err := h.Set(name, f); err == nil {
			h.funcs[name] = f
		} else if h.verbose {
			l("Failed to register function for name '%s': %s", name, err)
		}

		h.lock.Unlock()
	}
}

// UnregisterFunctions releases registered functions.
func (h *WasmHelper) UnregisterFunctions(names []string) {
	if h.verbose {
		prettified, _ := Prettify(names)

		l("Unregistering functions with names: %s", prettified)
	}

	for _, name := range names {
		h.lock.Lock()

		if fn, err := h.Get(name); err == nil {
			if fn.Type() == js.TypeFunction {
				if f, exists := h.funcs[name]; exists {
					f.Release()

					delete(h.funcs, name)
				} else {
					l("Failed to get registered function for name '%s'", name)
				}
			} else {
				l("Retrieved value '%s' is not a function.", name)
			}
		} else if h.verbose {
			l("Failed to get function for name '%s': %s", name, err)
		}

		h.lock.Unlock()
	}
}

// Wait blocks until stopped manually, for long-running routines.
//
// May panic when there is no registered callback or event listener.
func (h *WasmHelper) Wait() {
	if h.verbose {
		l("Waiting...")
	}

	// wait...
	<-h.block

	if h.verbose {
		l("Stopped waiting")
	}
}

// Stop stops blocking.
func (h *WasmHelper) Stop() {
	if h.verbose {
		l("Stopping waiting...")
	}

	h.block <- struct{}{}
}

// Get retrieves value for given name.
//
// (eg: name = "document.someparent.somechild.value")
func (h *WasmHelper) Get(name string) (js.Value, error) {
	if h.verbose {
		l("Getting value for name: '%s'", name)
	}

	names := strings.Split(name, ".")
	count := len(names)

	if count > 0 {
		value, _ := h.get(js.Null(), names)

		if h.verbose {
			l("Got value: %+v for name: '%s'", value, name)
		}

		return value, nil
	}

	return js.Undefined(), fmt.Errorf("could not get value, given name is empty")
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
			l("Parent not given, using global for names: %s", prettified)
		}

		parent = js.Global()
	}

	// child
	child := parent.Get(names[0])
	if child.IsUndefined() {
		l("Error: '%s' is undefined", names[0])

		return child, nil
	} else if child.IsNull() {
		l("Error: '%s' is null", names[0])

		return child, nil
	}

	if h.verbose {
		prettified, _ := Prettify(names[1:])
		l("Recursing on child: %+v with names: %s", child, prettified)
	}

	// recurse
	return h.get(child, names[1:])
}

// Set sets value for given name. (eg: 'document.someparent.somechild.value')
func (h *WasmHelper) Set(name string, value any) error {
	if h.verbose {
		l("Setting value: %+v for name: '%s'", value, name)
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
			return fmt.Errorf("could not set value, '%s' is undefined or null", strings.Join(parentNames, "."))
		}

		lastName = names[count-1]
	} else if count == 1 {
		parent = js.Global()
		lastName = names[0]
	} else {
		return fmt.Errorf("could not set value, given name is empty")
	}

	// set value
	parent.Set(lastName, value)

	return nil
}

// SetOn sets value for given property name on given object.
func (h *WasmHelper) SetOn(obj js.Value, propertyName string, value any) error {
	if h.verbose {
		l("Setting value: %+v on %+v for name: '%s'", value, obj, propertyName)
	}

	// undefined / null check
	if obj.IsUndefined() || obj.IsNull() {
		return fmt.Errorf("could not set value: '%+v' for name: '%s' on object which is undefined or null", value, propertyName)
	}

	obj.Set(propertyName, value)

	return nil
}

// Call calls a function with given name and arguments.
func (h *WasmHelper) Call(name string, args ...any) (js.Value, error) {
	if h.verbose {
		prettified, _ := Prettify(args)
		l("Calling '%s' with arguments: %s", name, prettified)
	}

	names := strings.Split(name, ".")
	parentNames := names[:len(names)-1]
	funcName := names[len(names)-1]

	var parent js.Value
	if len(names) >= 2 {
		parent, _ = h.Get(strings.Join(parentNames, "."))
	} else {
		parent = js.Global()
	}

	// undefined / null check
	if parent.IsUndefined() || parent.IsNull() {
		return parent, fmt.Errorf("could not call: '%s' on a parent which is undefined or null", name)
	}

	function := parent.Get(funcName)

	// undefined / null check
	if function.IsUndefined() || function.IsNull() {
		return function, fmt.Errorf("could not call: '%s' which is undefined or null", funcName)
	}

	// type check
	if function.Type() != js.TypeFunction {
		return js.Undefined(), fmt.Errorf("could not call '%s' which is not a function", name)
	}

	if h.verbose {
		prettified, _ := Prettify(args)
		l("Calling '%s' on %+v with arguments: %s", funcName, parent, prettified)
	}

	return parent.Call(funcName, args...), nil
}

// CallOn calls a function on a object with given name and arguments.
func (h *WasmHelper) CallOn(obj js.Value, funcName string, args ...any) (js.Value, error) {
	if h.verbose {
		prettified, _ := Prettify(args)
		l("Calling '%s' on %+v with arguments: %s", funcName, obj, prettified)
	}

	if obj.IsUndefined() || obj.IsNull() {
		return obj, fmt.Errorf("could not call: '%s' on an object which is undefined or null", funcName)
	}

	function := obj.Get(funcName)

	// undefined / null check
	if function.IsUndefined() || function.IsNull() {
		return function, fmt.Errorf("could not call '%s' on %+v which is undefined or null", funcName, obj)
	}

	// type check
	if function.Type() != js.TypeFunction {
		return js.Undefined(), fmt.Errorf("could not call '%s' on %+v which is not a function", funcName, obj)
	}

	if h.verbose {
		prettified, _ := Prettify(args)
		l("Calling '%s' on %+v with arguments: %s", funcName, obj, prettified)
	}

	return obj.Call(funcName, args...), nil
}

// Invoke invokes given function with arguments.
func (h *WasmHelper) Invoke(function js.Value, args ...any) (js.Value, error) {
	if h.verbose {
		prettified, _ := Prettify(args)
		l("Invoking %+v with arguments: %s", function, prettified)
	}

	// undefined / null check
	if function.IsUndefined() || function.IsNull() {
		return function, fmt.Errorf("could not invoke %+v which is undefined or null", function)
	}

	// type check
	if function.Type() != js.TypeFunction {
		return js.Undefined(), fmt.Errorf("could not invoke %+v which is not a function", function)
	}

	if h.verbose {
		prettified, _ := Prettify(args)

		l("Invoking %+v arguments: %s", function, prettified)
	}

	return function.Invoke(args...), nil
}

// print log to the console
func l(format string, v ...any) {
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
		return fmt.Sprintf("%+v", value), fmt.Errorf("failed to marshal given value: %s", err)
	}

	return string(bytes), nil
}
