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

	var nickClient string

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

			// handler - REGISTER-ATUADOR
		case "REGISTER-ATUADOR":
			nick := strings.TrimSpace(parts[1])
			resp := repository.SalvarAtuador(nick, conn, leitor)
			log.Println(resp)

			// Espera sinal de desconexão sem tocar na conexão
			a := repository.GetAtuador(nick)
			if a != nil {
				<-a.Done // bloqueia aqui até alguém fechar o canal
			}
			return

		case "REGISTER-CLIENT":
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
