package handler

import (
	"broker/repository"
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
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

			a := repository.GetAtuador(nick)
			if a != nil {
				go func() {
					for {
						a.Mu.Lock()
						buf := make([]byte, 1)
						conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
						_, err := conn.Read(buf)
						conn.SetReadDeadline(time.Time{})
						a.Mu.Unlock()

						if err != nil {
							// verifica se é timeout (normal) ou desconexão real
							if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
								continue // timeout normal, atuador ainda vivo
							}
							// erro real = atuador caiu
							log.Printf("Atuador %s caiu\n", nick)
							select {
							case <-a.Done:
							default:
								close(a.Done)
							}
							return
						}
					}
				}()

				<-a.Done
			}

			repository.RemoverAtuador(nick)
			log.Printf("Atuador %s DESCONECTADO\n", nick)
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

		case "HELP":
			help := "============================= HELP =============================\n" +
				"REGISTER-CLIENT <nick>              → registra o cliente\n" +
				"REGISTER-ATUADOR <nick>             → registra o atuador\n" +
				"COMMAND <nickAtuador> <ON/OFF>       → envia comando ao atuador\n" +
				"SEGUIR-SENSOR <nickSensor>           → inscreve no sensor\n" +
				"PARAR-SENSOR <nickSensor>            → cancela inscrição no sensor\n" +
				"LIST-SENSORES                        → lista sensores conectados\n" +
				"LIST-ATUADORES                       → lista atuadores conectados\n" +
				"LIST-CLIENTES                        → lista clientes conectados\n" +
				"PING                                 → verifica conexão com o broker\n" +
				"QUIT                                 → encerra a conexão\n" +
				"================================================================\n"
			conn.Write([]byte(help))

		default:
			conn.Write([]byte("Comando inválido, use HELP\n"))
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
