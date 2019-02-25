package main

import . "github.com/davidkhala/goutils"

func panicEcosystem(message string) {
	PanicString("ECOSYSTEM|" + message)
}
