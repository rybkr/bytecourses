package main

import "syscall/js"

func initAuth() {
	js.Global().Get("document").
		Call("getElementById", "login-btn").
		Call("addEventListener", "click", js.FuncOf(login))
}

func login(js.Value, []js.Value) any {
	return nil
}
