package main

import (
	"sync"

	"./resync"
)

func main() {
	var wg sync.WaitGroup

	register := resync.New()
	go register.Read(wg)
	register.Update()
}
