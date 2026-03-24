package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"os"
)

const (
	SERVER_ADDR = "192.168.0.103:9091" // IP do servidor TCP
	NICK = "atuador1"
)

func getServerAddr() string {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "localhost:9091" // fallback
	}
	return addr
}

func main() {
	//conecta no servidor
	conn, err := net.Dial("tcp", getServerAddr())
	if err != nil {
		log.Fatal("Erro ao conectar no servidor:", err)
	}
	defer conn.Close()

	log.Println("Conectado ao servidor:", SERVER_ADDR)

	// se registrar nem precisa
	conn.Write([]byte("REGISTER-ATUADOR " + NICK + "\n"))

	reader := bufio.NewReader(conn)

	for {
		//espera comando
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Conexão encerrada pelo servidor")
			return
		}

		msg = strings.TrimSpace(msg)

		if msg == "" {
			continue
		}

		log.Println("Comando recebido:", msg)


		// executa comando
		resposta := executarComando(msg)

		fmt.Println(resposta)
		//responde (opcional)
		conn.Write([]byte(resposta + "\n"))
	}
}

func executarComando(cmd string) string {
	parts := strings.Split(cmd, " ")

	switch parts[0] {

	case "ON":
		fmt.Println("🔵 Atuador LIGADO")
		return "OK ON"

	case "OFF":
		fmt.Println("⚫ Atuador DESLIGADO")
		return "OK OFF"

	case "SET":
		if len(parts) < 2 {
			return "ERRO SET"
		}
		valor := parts[1]
		fmt.Println("⚙️ Ajustando valor para:", valor)
		return "OK SET " + valor

	default:
		return "COMANDO DESCONHECIDO"
	}
}