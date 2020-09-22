package jsonrpc_server

import (
	"github.com/DizoftTeam/jsonrpc_server/utils"

	"github.com/mitchellh/mapstructure"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

const (
	rpcVersion = "2.0" // Supported protocol version
)

// Array of methods
var methods = map[string]Method{}

// RPCRequest RPC struct
type RPCRequest struct {
	Version string      `mstruct:"jsonrpc"` // Protocol version
	Method  string      `mstruct:"method"`  // Method name
	Params  interface{} `mstruct:"params"`  // Method params
	ID      int         `mstruct:"id"`      // Request id
}

// RPCError Error struct
type RPCError struct {
	Code    int    // Error code
	Message string // Error message
}

// Method Alias on func
type Method func(params interface{}) (interface{}, *RPCError)

// RPCMethod Interface for struct style method
type RPCMethod interface {
	Handler(params interface{}) (interface{}, *RPCError)
}

// ----------- STRUCT METHODS -----------

// Register Add method based on struct style
func Register(name string, handler RPCMethod) {
	RegisterFunc(name, handler.Handler)
}

// RegisterFunc Add method based on lambda func
func RegisterFunc(name string, method Method) {
	methods[name] = method

	log.Printf("Register method: %v\n", name)
}

// --------------- PUBLIC ---------------

// Handler Main point function
func Handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Add("Content-Type", "application/json")

	// TIP: CORS
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "POST")

	var request interface{}
	var response interface{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_, _ = fmt.Fprintf(w, `{"jsonrpc": "2.0", "error": {"code": -42700, "message": "Common Error"}}`)

		return
	}

	reqType := reflect.ValueOf(request).Kind()

	if reqType == reflect.Slice {
		var responses []string

		for _, item := range request.([]interface{}) {
			responses = append(responses, processRequest(item.(map[string]interface{})))
		}

		// TODO: maybe it can make more beautiful
		response = "[" + strings.Join(responses, ",") + "]"
	} else {
		response = processRequest(request.(map[string]interface{}))
	}

	_, _ = fmt.Fprint(w, response)
}

// EmptyRequestError Wrong data or Empty request
// Helper function for most cases
func EmptyRequestError() (interface{}, *RPCError) {
	return nil, &RPCError{
		Code:    -20,
		Message: "Wrong data or Empty request",
	}
}

// --------------- PRIVATE ---------------

func processRequest(request utils.Object) string {
	req := RPCRequest{}

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &req,
		TagName: "mstruct",
	})

	// Check wrong JSON data
	if err := decoder.Decode(request); err != nil {
		return performError(req, -32700, "Parse error")
	}

	// Check RPC format
	if req.Version != rpcVersion {
		return performError(req, -32600, "Invalid Request")
	}

	// If method not found
	if _, ok := methods[req.Method]; !ok {
		return performError(req, -32601, "Method not found")
	}

	// Run method
	method := methods[req.Method]
	result, rpcError := method(req.Params)

	// Check is not notify type
	if req.ID != 0 {
		if rpcError != nil {
			return performError(req, rpcError.Code, rpcError.Message)
		}

		return performSuccess(req, result)
	}

	return ""
}

// performError Format success response
func performSuccess(rpc RPCRequest, data interface{}) string {
	return performResponse(rpc, "result", data)
}

// performError Format error response
func performError(rpc RPCRequest, code int, message string) string {
	return performResponse(rpc, "error", utils.Object{
		"code":    code,
		"message": message,
	})
}

// performResponse Create response
func performResponse(rpc RPCRequest, key string, value interface{}) string {
	_struct := utils.Object{
		"jsonrpc": "2.0",
		"id":      rpc.ID,
		key:       value,
	}

	result, _ := json.Marshal(_struct)

	return string(result)
}
