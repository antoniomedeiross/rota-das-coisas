package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	SERVER_ADDR = "192.168.0.103:9091" // IP do servidor TCP
	NICK        = "atuador1"
)

var (
	ativo = false
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

		fmt.Println("---------- CONEXÃO ESTABELECIDA COM O SERVIDOR ----------")
		fmt.Println("AGUARDANDO COMANDOS...")
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
		if ativo == true {
			fmt.Println("Atuador JÁ ESTÁ LIGADO")
			return "ATUADOR JA ESTA LIGADO"
		} else if ativo == false {
			fmt.Println("Atuador LIGADO")
			ativo = true
			return "ATUADOR LIGADO"
		}
		return "OCORREU UM ERRO"

	case "OFF":
		if ativo == true {
			fmt.Println("Atuador DESLIGADO")
			ativo = false
			return "ATUADOR DESLIGADO"
		} else if ativo == false {
			fmt.Println("Atuador JA ESTÁ DESLIGADO")
			ativo = true
			return "ATUADOR JA ESTA DESLIGADO"
		}
		return "OCORREU UM ERRO"
	default:
		return "COMANDO DESCONHECIDO"
	}
}
