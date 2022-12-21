# Go JsonRpc Server

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/DizoftTeam/jsonrpc_server.svg)](https://github.com/DizoftTeam/jsonrpc_server)
[![GitHub tag](https://img.shields.io/github/tag/DizoftTeam/jsonrpc_server.svg)](https://GitHub.com/DizoftTeam/jsonrpc_server/tags/)

[![GitHub stars](https://img.shields.io/github/stars/DizoftTeam/jsonrpc_server.svg?style=social&label=Star&maxAge=2592000)](https://GitHub.com/DizoftTeam/jsonrpc_server/stargazers/)
[![GitHub issues](https://img.shields.io/github/issues/DizoftTeam/jsonrpc_server.svg)](https://GitHub.com/DizoftTeam/jsonrpc_server/issues/)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)
[![Open Source? Yes!](https://badgen.net/badge/Open%20Source%20%3F/Yes%21/blue?icon=github)](https://github.com/Naereen/badges/)

Base implementation of JSONRpc v2.0 in Go

## Breaking changes of v2

### HttpHandler

Cause: add new handler

> Old

http.HandleFunc("/", jsonrpc.Handler)

> New

http.HandleFunc("/", jsonrpc.HttpHandler)

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

// UserLogin controller for login method
type UserLogin struct {}

// Handler worker
func (u UserLogin) Handler(params interface{}) (interface{}, *jsonrpc.RPCError) {
    // Some logic/magic here.
    // It's like a controller
    
    // To getting raw request you can run this
    // session := jsonrpc.NewSession()

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

    http.HandleFunc("/", jsonrpc.HttpHandler)

    log.Print("\nStarting server at :8089\n")

    if err := http.ListenAndServe(":8089", nil); err != nil {
      log.Panic("Cant start server", err)
    }
}
```
