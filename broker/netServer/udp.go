package server

import(
	"log"
	"net"
	"broker/handler"
	"broker/config"
)
// vai iniciar o canal UDP e chamar os handlers

func StartUdp() {
	addr := &net.UDPAddr{IP: net.ParseIP(config.UDP_IP), Port: config.UDP_PORT}

	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		log.Println("ERRO AO ABRIR PORTA DO SERVIDOR")
	}

	log.Println("SERVER INICIADO NO ENDEREÇO --> ", addr)

	for {
		buffer := make([]byte, 1024)
		//client = todo mundo que mandar requisições pro server
		n, clientAddr, err := conn.ReadFromUDP(buffer)

		if err != nil {
			log.Println("ERRO AO RECEBER LER MENSAGEM VINDA DO CANAL UDP")
			continue
		}

		data := make([]byte, n)

		//copiando senao vai dar merda no buffer
		copy(data, buffer[:n])

		// processamento é feito em paralelo
		go handler.HandleRequestUdp(data, clientAddr, conn)
	}
}