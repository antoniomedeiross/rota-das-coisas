package handler

import (
	"strings"
	"broker/repository"
	"log"
	"net"
)
// go rotine que processa as requisicoes do canal UDP

func HandleRequestUdp(data []byte, clientAddr *net.UDPAddr, conn *net.UDPConn) {

	msg := strings.SplitN(string(data), " ", 2)

	switch msg[0] {
	case "REGISTER-SENSOR":
		resp := repository.SalvarSensor(msg[1], clientAddr)
		conn.WriteToUDP([]byte(resp), clientAddr)

		log.Println("DISPOSITIVO CONECTADO", msg[1])
		
	case "DATA":
		repository.EnviarDados(msg[1], conn, clientAddr)

	case "HELP":
		repository.MenuHelp(conn, clientAddr)

	case "DEBUG":
		//fmt.Println(repository.Dispositivos)
		//log.Println(repository.Atuadores)
		//fmt.Println(repository.Clientes)
		log.Println("SENSORES CONECTADOS =", len(repository.Dispositivos))
		log.Println("CLIENTES CONECTADOS =", len(repository.Clientes))

	default:
		log.Println("COMANDO INVÁLIDO")
		conn.WriteToUDP([]byte("COMANDO INVÁLIDO, USE HELP\n"), clientAddr)
	}

	
}