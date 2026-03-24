package server


import (
	"log"
	"net"
	"broker/handler"
)

// vai iniciar o canal TCP e chamar os handlers


func StartTCP() {
	listener, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("TCP rodando em :9091")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handler.HandleRequestTcp(conn)
	}
}