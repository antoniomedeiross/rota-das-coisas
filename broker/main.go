package main

import (
	"broker/netServer"
	"broker/repository"
)



func main() {
	go server.StartTCP()
	go server.StartUdp()

	go repository.HeartBeat()
	
	select {} // nao deixa a main parar
}







