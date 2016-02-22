package sherpa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

// Sherpa version this package implements. Note that Sherpa is at version 0 and still in development and will probably change.
const SherpaVersion = 0

// handler that responds to all Sherpa-related requests.
type handler struct {
	baseURL   string
	id        string
	title     string
	version   string
	functions map[string]interface{}

	redirectURL string
	json        []byte
	javascript  []byte
}

// Documentation object, to be returned by a Sherpa API "_docs" function.
type Docs struct {
	Title     string          `json:"title"`
	Text      string          `json:"text"`
	Functions []*FunctionDocs `json:"functions"`
	Sections  []*Docs         `json:"sections"`
}

// Documentation for a single function Name.
// Text should be in markdown. The first line should be a synopsis showing parameters including types, and the return types.
type FunctionDocs struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// Sherpa API error object.
// Message is a human-readable error message.
// Code is optional, it can be used to handle errors programmatically.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// Sherpa API response type
type response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *Error      `json:"error,omitempty"`
}

func respond(w http.ResponseWriter, status int, r *response) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(r)
	if err != nil {
		log.Println("error writing json response:", err)
	}
}

// call function fn with a json body read from r.
// on success, the returned interface contains:
// - nil, if fn has no return value
// - single value, if fn had a single return value
// - slice of values, if fn had multiple return values
//
// on error, we always return an Error with the Code field set.
func call(fn interface{}, r io.Reader) (interface{}, *Error) {
	var request struct {
		Params json.RawMessage `json:"params"`
	}

	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, &Error{Code: SherpaBadRequest, Message: "invalid JSON request body: " + err.Error()}
	}

	t := reflect.TypeOf(fn)

	var params []interface{}
	err = json.Unmarshal(request.Params, &params)
	if err != nil {
		return nil, &Error{Code: SherpaBadRequest, Message: "invalid JSON request body: " + err.Error()}
	}

	need := t.NumIn()
	if t.IsVariadic() {
		need -= 1
		if len(params) < need {
			return nil, &Error{Code: SherpaBadParams, Message: fmt.Sprintf("bad number of parameters, got %d, want at least %d", len(params), need)}
		}
	} else {
		if len(params) != need {
			return nil, &Error{Code: SherpaBadParams, Message: fmt.Sprintf("bad number of parameters, got %d, want %d", len(params), need)}
		}
	}

	values := make([]reflect.Value, t.NumIn())
	args := make([]interface{}, t.NumIn())
	for i := range values {
		n := reflect.New(t.In(i))
		values[i] = n.Elem()
		args[i] = n.Interface()
	}

	err = json.Unmarshal(request.Params, &args)
	if err != nil {
		return nil, &Error{Code: SherpaBadParams, Message: fmt.Sprintf("could not parse parameters: " + err.Error())}
	}

	var results []reflect.Value
	if t.IsVariadic() {
		results = reflect.ValueOf(fn).CallSlice(values)
	} else {
		results = reflect.ValueOf(fn).Call(values)
	}
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0].Interface(), nil
	default:
		rr := make([]interface{}, len(results))
		for i, v := range results {
			rr[i] = v.Interface()
		}
		return rr, nil
	}
}

// NewHandler returns a new http.Handler that serves all Sherpa API-related requests.
//
// baseURL must be the URL this API is available at.
// id is the variable name for the API object the JavaScript client library.
// title should be a human-readable name of the API.
// functions are the functions you want to make available through this handler.
//
// This handler expects to be called with any path elements stripped using http.StripPrefix.
func NewHandler(baseURL, id, title, version string, functions map[string]interface{}) (http.Handler, error) {
	var redirectURL string
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http", "https":
		redirectURL = fmt.Sprintf("%s://sherpa.irias.nl/x/%s%s", u.Scheme, u.Host, u.EscapedPath())
	default:
		return nil, fmt.Errorf("Unsupported URL scheme %#v", u.Scheme)
	}

	names := make([]string, 0, len(functions))
	for name, fn := range functions {
		if reflect.TypeOf(fn).Kind() != reflect.Func {
			return nil, fmt.Errorf("sherpa handler: %#v is not of type function", name)
		}
		names = append(names, name)
	}

	marshal := func(v interface{}) []byte {
		buf, err := json.Marshal(v)
		if err != nil {
			log.Panicf("marshal json %#v: %s", v, err)
		}
		return buf
	}

	xjson := marshal(map[string]interface{}{
		"id":            id,
		"title":         title,
		"functions":     names,
		"baseurl":       baseURL,
		"version":       version,
		"sherpaVersion": SherpaVersion,
	})

	js := sherpaJS
	js = bytes.Replace(js, []byte("ID"), marshal(id), -1)
	js = bytes.Replace(js, []byte("TITLE"), marshal(title), -1)
	js = bytes.Replace(js, []byte("VERSION"), marshal(version), -1)
	js = bytes.Replace(js, []byte("URL"), marshal(baseURL), -1)
	js = bytes.Replace(js, []byte("FUNCTIONS"), marshal(names), -1)

	h := &handler{
		baseURL:     baseURL,
		id:          id,
		title:       title,
		functions:   functions,
		version:     version,
		redirectURL: redirectURL,
		json:        xjson,
		javascript:  js}
	return h, nil
}

func badMethod(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// Serve a HTTP request for this Sherpa API.
// ServeHTTP expects the request path is stripped from the path it was mounted at with the http package.
//
// The following endpoints are handled:
// 	- sherpa.json, describing this API.
// 	- sherpa.js, a small stand-alone client JavaScript library that makes it trivial to start using this API from a browser.
// 	- functionName, for function invocations on this API.
//
// HTTP response will have CORS-headers set, and support the OPTIONS HTTP method.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	hdr := w.Header()
	hdr.Set("Access-Control-Allow-Origin", "*")
	hdr.Set("Access-Control-Allow-Methods", "GET, POST")
	hdr.Set("Access-Control-Allow-Headers", "Content-Type")

	switch {
	case r.URL.Path == "":
		http.Redirect(w, r, h.redirectURL, 302)

	case r.URL.Path == "sherpa.json":
		switch r.Method {
		case "OPTIONS":
			w.WriteHeader(204)
		case "GET":
			hdr.Set("Content-Type", "application/json; charset=utf-8")
			hdr.Set("Cache-Control", "no-cache")
			_, err := w.Write(h.json)
			if err != nil {
				log.Println("writing sherpa.json response:", err)
			}
		default:
			badMethod(w)
		}

	case r.URL.Path == "sherpa.js":
		if r.Method != "GET" {
			badMethod(w)
			return
		}
		hdr.Set("Content-Type", "text/javascript; charset=utf-8")
		hdr.Set("Cache-Control", "no-cache")
		_, err := w.Write(h.javascript)
		if err != nil {
			log.Println("writing sherpa.js response:", err)
		}

	default:
		name := r.URL.Path
		fn, ok := h.functions[name]
		switch r.Method {
		case "OPTIONS":
			w.WriteHeader(204)

		case "POST":
			// xxx check file upload

			if !ok {
				respond(w, 404, &response{Error: &Error{Code: SherpaBadFunction, Message: fmt.Sprintf("function %v does not exist", name)}})
				return
			}

			r, xerr := call(fn, r.Body)
			if xerr != nil {
				respond(w, 200, &response{Error: xerr})
			} else {
				respond(w, 200, &response{Result: r})
			}

		case "GET":
			badMethod(w)
			// xxx parse params, call function, return jsonp

		default:
			badMethod(w)
		}
	}
}
