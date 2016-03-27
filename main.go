package main

import "./register"

func main() {
	reg := register.New()

	go reg.Read()
	reg.Update()
}
