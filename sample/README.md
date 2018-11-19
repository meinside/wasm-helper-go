# wasm-helper-go/sample

Sample codes for wasm-helper.

## Contents

- main.go: source code of the sample application
- index.html: html file that will load the compiled wasm file
- wasm_exec.js: wasm helper script copied from `$GOROOT/misc/wasm/wasm_exec.js`

## Build

```bash
$ GOOS=js GOARCH=wasm go build -o sample.wasm main.go
```

## Run

Run any http server on this directory and open `index.html`.

```bash
# for example, start a simple webserver with ruby on port 8888,
$ ruby -rwebrick -e's=WEBrick::HTTPServer.new(Port:8888,DocumentRoot:Dir.pwd);trap("INT"){s.shutdown};s.start'

# and open the index.html file
$ open http://localhost:8888
```

