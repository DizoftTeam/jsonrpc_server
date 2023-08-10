package jsonrpc_server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/DizoftTeam/jsonrpc_server/utils"
	"github.com/mitchellh/mapstructure"
)

const (
	rpcVersion = "2.0" // Supported protocol version

	notifyResponse = "JsonRpc_Notify_Response"
)

var (
	methods     = map[string]Method{} // Array of methods
	httpRequest *http.Request         // Current request

	jlog = log.New(os.Stderr, "[JSONRpc]", log.LstdFlags)
)

// RPCRequest RPC struct
type RPCRequest struct {
	Version string `mstruct:"jsonrpc"` // Protocol version
	Method  string `mstruct:"method"`  // Method name
	Params  any    `mstruct:"params"`  // Method params
	ID      int    `mstruct:"id"`      // Request id
}

// RPCError Error struct
type RPCError struct {
	Code    int    // Error code
	Message string // Error message
}

// Method Alias on func
type Method func(params any) (any, *RPCError)

// RPCMethod Interface for struct style method
type RPCMethod interface {
	Handler(params any) (any, *RPCError)
}

// Session contains request info
type Session struct {
	Request *http.Request // Current request
}

// ----------- STRUCT METHODS -----------

// Register Add method based on struct style
func Register(name string, handler RPCMethod) {
	RegisterFunc(name, handler.Handler)
}

// RegisterFunc Add method based on lambda func
func RegisterFunc(name string, method Method) {
	methods[name] = method

	jlog.Printf("Register method: %v\n", name)
}

// NewSession create new request session
func NewSession() *Session {
	return &Session{
		Request: httpRequest,
	}
}

// --------------- PUBLIC ---------------

// CustomHandler used for custom incoming messages point (like WS)
func CustomHandler(rawJsonData []byte) string {
	var request any
	var response string

	if err := json.Unmarshal(rawJsonData, &request); err != nil {
		return `{"jsonrpc": "2.0", "error": {"code": -42700, "message": "Common Error"}}`
	}

	reqType := reflect.ValueOf(request).Kind()

	if reqType == reflect.Slice {
		var responses []string

		for _, item := range request.([]any) {
			r := processRequest(item.(map[string]any))

			if r != notifyResponse {
				responses = append(responses, r)
			}
		}

		var rr []string
		for _, r := range responses {
			if r != "" && r != notifyResponse {
				rr = append(rr, r)
			}
		}

		response = fmt.Sprintf("[%s]", strings.Join(rr, ","))
	} else {
		response = processRequest(request.(map[string]any))

		if response == notifyResponse {
			response = ""
		}
	}

	return response
}

// HttpHandler Main point function of http handler
func HttpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Add("Content-Type", "application/json")

	// TIP: CORS
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "POST")

	httpRequest = r

	var request any
	var response any

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_, _ = fmt.Fprintf(w, `{"jsonrpc": "2.0", "error": {"code": -42700, "message": "Common Error"}}`)

		return
	}

	reqType := reflect.ValueOf(request).Kind()

	if reqType == reflect.Slice {
		var responses []string

		for _, item := range request.([]any) {
			pr := processRequest(item.(map[string]any))

			if pr != notifyResponse {
				responses = append(responses, pr)
			}
		}

		var rr []string
		for _, r := range responses {
			if r != "" {
				rr = append(rr, r)
			}
		}

		// TODO: maybe it can make more beautiful
		response = "[" + strings.Join(rr, ",") + "]"
	} else {
		response = processRequest(request.(map[string]any))
	}

	_, _ = fmt.Fprint(w, response)

	httpRequest = nil
}

// EmptyRequestError Wrong data or Empty request
// Helper function for most cases
func EmptyRequestError() (any, *RPCError) {
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

	return notifyResponse
}

// performError Format success response
func performSuccess(rpc RPCRequest, data any) string {
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
func performResponse(rpc RPCRequest, key string, value any) string {
	_struct := utils.Object{
		"jsonrpc": "2.0",
		"id":      rpc.ID,
		key:       value,
	}

	result, _ := json.Marshal(_struct)

	return string(result)
}
