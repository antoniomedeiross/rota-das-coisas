package handler

import (
	"broker/repository"
	"bufio"
	"log"
	"net"
	"strings"
)

// goroutine que processa o canal TCP

func HandleRequestTcp(conn net.Conn) {
	defer conn.Close() // no fim fecha a conexao

	addr := conn.RemoteAddr().String()
	log.Println("Cliente TCP conectado:", addr)

	leitor := bufio.NewReader(conn)

	for {
		//lê mensagem linha a linha
		msg, err := leitor.ReadString('\n')
		if err != nil {
			log.Println("Cliente desconectado:", addr)
			return
		}

		msg = strings.TrimSpace(msg)

		if msg == "" {
			continue
		}

		log.Println("Recebido:", msg)

		
		// parse comando
		parts := strings.SplitN(msg, " ", 2)

		switch parts[0] {
			case "COMMAND":
				if len(parts) < 2 {
					conn.Write([]byte("Uso: COMMAND <nickAtuador> <ação>\n"))
					continue
				}
				handleCommand(parts[1], conn)

			case "REGISTER-ATUADOR":

				if strings.HasPrefix(msg, "REGISTER-ATUADOR") {
					nick := strings.TrimSpace(parts[1])
					repository.SalvarAtuador(nick, conn)

					log.Println("Atuador registrado:", nick)

					//entra em modo passivo não lê mais comandos
					select {}
				}
		

			case "REGISTER-CLIENT":
				nick := parts[1]
				repository.SalvarCliente(nick, conn)
	
			
			case "QUIT":
				conn.Write([]byte("Conexão encerrada\n"))
				return

			case "PING":
				conn.Write([]byte("PONG\n"))

			default:
				conn.Write([]byte("Comando inválido\n"))
		}
	}
}

func handleCommand(cmd string, conn net.Conn) {
	parts := strings.SplitN(cmd, " ", 2)

	if len(parts) < 2 {
		conn.Write([]byte("Uso: COMMAND <nick> <acao>\n"))
		return
	}

	nick := strings.TrimSpace(parts[0])
	acao := strings.TrimSpace(parts[1])

	resp := repository.ComandarAtuador(nick, acao)

	log.Println("Resposta do atuador:", resp)

	conn.Write([]byte(resp))
}