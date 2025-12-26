package main

import "syscall/js"

func main() {
	path := js.Global().Get("location").Get("pathname").String()

	switch path {
	case "/login":
		initAuth()
	}

	select {}
}
