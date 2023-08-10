package main

import "example.com/m/v2/server"

func main() {
	servers := server.NewServer("127.0.0.1", 8888)
	servers.Start()
}
