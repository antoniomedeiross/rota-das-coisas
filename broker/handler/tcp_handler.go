package handler

import (
	"broker/repository"
	"bufio"
	"fmt"
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

	var nickClient string

	for {
		//lê mensagem linha a linha
		msg, err := leitor.ReadString('\n')
		if err != nil {
			log.Println("Cliente desconectado:", addr)
			repository.RemoverClient(nickClient)
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

			// handler - REGISTER-ATUADOR
		case "REGISTER-ATUADOR":
			nick := strings.TrimSpace(parts[1])
			resp := repository.SalvarAtuador(nick, conn, leitor)
			log.Println(resp)

			// Espera sinal de desconexão
			a := repository.GetAtuador(nick)
			if a != nil {
				go func() {
					buf := make([]byte, 1)
					for {
						_, err := conn.Read(buf)
						if err != nil {
							log.Printf("Atuador %s caiu: %v\n", nick, err)
							select {
							case <-a.Done:
							default:
								close(a.Done)
							}
							return
						}
					}
				}()

				<-a.Done // bloqueia aqui
			}

			repository.RemoverAtuador(nick) // remove da lista
			log.Printf("Atuador %s DESCONECTADO \n", nick)

			return

		case "REGISTER-CLIENT":
			if len(parts) < 2 {
				conn.Write([]byte("ERRO: nick não informado\n"))
				continue
			}
			nick := parts[1]
			resp := repository.SalvarCliente(nick, conn)

			// so pode acontecer se salvar direito...
			nickClient = nick

			conn.Write([]byte(resp))
		case "SEGUIR-SENSOR":
			nickSensor := parts[1]
			var resp string
			if nickClient != "" {
				resp = repository.SeguirSensor(nickSensor, nickClient)
			} else {
				resp = "VOCE PRECISA SE REGISTRAR\n"
			}
			conn.Write([]byte(resp))

		case "LIST-SENSORES":
			repository.ListarDispositivosConectados(conn)
		case "LIST-ATUADORES":
			repository.ListarAtuadoresConectados(conn)

		case "LIST-CLIENTES":
			repository.ListarClientesConectados(conn)
		case "PARAR-SENSOR":
			nickSensor := parts[1]
			var resp string
			if nickClient != "" {
				resp = repository.PararSensor(nickSensor, nickClient)
			} else {
				resp = "VOCE PRECISA SE REGISTRAR\n"
			}
			conn.Write([]byte(resp))

		case "QUIT":
			conn.Write([]byte("Conexão encerrada\n"))
			repository.RemoverClient(nickClient)
			fmt.Println("Client " + nickClient + " desconectado")
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

	go func() {
		resp := repository.ComandarAtuador(nick, acao)

		log.Println("Resposta do atuador:", resp)

		conn.Write([]byte(resp))
	}()

}
