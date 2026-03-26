package main

import (
	"broker/netServer"
	"broker/repository"
)

/*
	dar 2 REGISTER salva 2 vezes o msm adr com nick diferentes

	Se tiver muitos sensores:
		1000 sensores → 1000 goroutines por segundo pode virar problema
		Por enquanto Pode ignorar isso Mas saiba que existe solução {worker pool}

*/

func main() {
	go server.StartTCP()
	go server.StartUdp()

	go repository.HeartBeat()
	
	select {} // nao deixa a main parar
}







