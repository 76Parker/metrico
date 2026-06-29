package main

import "runtime"

func main() {

	metrics := runtime.MemStats{}
	runtime.ReadMemStats(&metrics)

}
