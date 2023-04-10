package main

import (
	"fmt"
	"syscall/js"
)

func getData(url string, params map[string]string) (string, error) {
	done := make(chan struct{}, 0)
	jsParams := js.Global().Get("Object").New()
	for k, v := range params {
		jsParams.Set(k, v)
	}

	var data string
	var err error
	js.Global().Get("$").Call("get", url, jsParams, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer close(done)
		if len(args) > 0 {
			response := args[0]
			data = js.Global().Get("JSON").Call("stringify", response).String()
			return nil
		}
		err = fmt.Errorf("No data returned from $.get")
		return nil
	}))
	<-done

	return data, err
}

func postData(url string, params map[string]string, headers map[string]string) (string, error) {
	done := make(chan struct{}, 0)
	jsParams := js.Global().Get("Object").New()
	for k, v := range params {
		jsParams.Set(k, v)
	}

	jsHeaders := js.Global().Get("Object").New()
	for k, v := range headers {
		jsHeaders.Set(k, v)
	}

	obj := js.Global().Get("Object").New()
	obj.Set("type", "GET")
	obj.Set("url", url)
	obj.Set("data", jsParams)
	obj.Set("headers", jsHeaders)

	var data string
	var err error
	obj.Set("success", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer close(done)
		if len(args) > 0 {
			response := args[0]
			data = js.Global().Get("JSON").Call("stringify", response).String()
			return nil
		}
		err = fmt.Errorf("No data returned from $.get")
		return nil
	}))

	js.Global().Get("$").Call("get", obj)
	<-done

	return data, err
}
