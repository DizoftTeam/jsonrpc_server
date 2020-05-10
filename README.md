# Go JsonRpc Server

Base implementation of JSONRpc in Go

## How to get

Use follow command

```bash
go get github.com/DizoftTeam/jsonrpc_server
```

## Example

Command example below

```go
package main

import (
    jsonrpc "github.com/DizoftTeam/jsonrpc_server"

    "fmt"
    "log"
    "net/http"
)

type UserLogin struct {}

func (u UserLogin) Handler(params interface{}) (interface{}, *jsonrpc.RPCError) {
    // Some logic here.
    // It's like a controller
    
    // Success
    return "Login ok!", nil 

    // Fail
    //return nil, &jsonrpc.RPCError{
    //    Code: -10,
    //    Message: "Cant login",
    //}
}

// Register methods and callbacks
func registerMethods() {
    jsonrpc.Register("user.login", UserLogin{})

    jsonrpc.RegisterFunc("version", func(params interface{}) (interface{}, *jsonrpc.RPCError) {
        return "1.0.0", nil
    })
}

func main() {
    registerMethods()

    http.HandleFunc("/", jsonrpc.Handler)

    fmt.Println()

    if err := http.ListenAndServe("8089", nil); err != nil {
        log.Panic("Cant start server", err)
    }
}
```
