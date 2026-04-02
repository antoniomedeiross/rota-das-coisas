package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const BROKER_ADDR = "192.168.0.103:9091"

var (
	conn               net.Conn
	leitor             *bufio.Reader
	nick               string
	sensorAtual        string
	aguardandoResposta = make(chan string) // sem buffer para sincronia exata
)

func getBrokerAddr() string {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		return BROKER_ADDR
	}
	return addr
}

func main() {
	var err error
	conn, err = net.Dial("tcp", getBrokerAddr())
	if err != nil {
		fmt.Println("Erro ao conectar no broker:", err)
		os.Exit(1)
	}
	defer conn.Close()

	leitor = bufio.NewReader(conn)
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Digite seu nick: ")
	scanner.Scan()
	nick = strings.TrimSpace(scanner.Text())

	// registro inicial síncrono
	enviar("REGISTER-CLIENT " + nick)
	resp, _ := leitor.ReadString('\n')
	if !strings.Contains(resp, "CONECTADO") {
		log.Fatalln("Erro ao registrar:", resp)
	}
	fmt.Println("\nConectado como:", nick)

	go escutarBroker()

	for {
		mostrarMenu()
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		opcao := strings.TrimSpace(scanner.Text())

		switch opcao {
		case "1":
			enviar("LIST-SENSORES")
			<-aguardandoResposta
		case "2":
			fmt.Print("Nick do sensor: ")
			scanner.Scan()
			ns := strings.TrimSpace(scanner.Text())
			seguirSensor(ns)
		case "3":
			pararSensor()
		case "4":
			enviar("LIST-ATUADORES")
			<-aguardandoResposta
		case "5":
			comandarAtuador(scanner)
		case "6":
			enviar("LIST-CLIENTES")
			<-aguardandoResposta
		case "7":
			enviar("QUIT")
			fmt.Println("Saindo...")
			os.Exit(0)
		default:
			fmt.Println("Opção inválida!")
		}
	}
}

func mostrarMenu() {
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
	fmt.Println("\n========================================")
	fmt.Println("         BROKER IoT - CLIENTE")
	fmt.Println("========================================")
	if sensorAtual != "" {
		fmt.Printf("  [SEGUINDO: %s]\n", sensorAtual)
		fmt.Println("========================================")
	}
	fmt.Println("  1 - Listar sensores")
	fmt.Println("  2 - Seguir sensor")
	fmt.Println("  3 - Parar de seguir sensor")
	fmt.Println("  4 - Listar atuadores")
	fmt.Println("  5 - Comandar atuador")
	fmt.Println("  6 - Listar clientes")
	fmt.Println("  7 - Sair")
	fmt.Println("========================================")
}

func seguirSensor(nickSensor string) {
	if sensorAtual != "" {
		fmt.Println("Já seguindo um sensor. Pare o atual primeiro (opção 3).")
		return
	}
	enviar("SEGUIR-SENSOR " + nickSensor)
	resp := <-aguardandoResposta
	if strings.Contains(strings.ToLower(resp), "ok") {
		sensorAtual = nickSensor
		fmt.Println("Seguindo agora:", nickSensor)
	} else {
		fmt.Println("Erro ao seguir sensor:", resp)
	}
}

func pararSensor() {
	if sensorAtual == "" {
		fmt.Println("Nenhum sensor ativo.")
		return
	}
	enviar("PARAR-SENSOR " + sensorAtual)
	<-aguardandoResposta
	sensorAtual = ""
	fmt.Println("\nInscrição cancelada.")
}

func comandarAtuador(scanner *bufio.Scanner) {
	fmt.Print("Nick do atuador: ")
	scanner.Scan()
	id := strings.TrimSpace(scanner.Text())

	fmt.Print("Comando (ON/OFF): ")
	scanner.Scan()
	cmd := strings.TrimSpace(scanner.Text())

	enviar(fmt.Sprintf("COMMAND %s %s", id, cmd))
	<-aguardandoResposta
}

func enviar(msg string) {
	conn.Write([]byte(msg + "\n"))
}


func escutarBroker() {
	for {
		msg, err := leitor.ReadString('\n')
		if err != nil {
			fmt.Println("\nConexão perdida.")
			os.Exit(0)
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}

		// dado do sensor sendo seguido → atualiza na mesma linha
		// dado do sensor sendo seguido → atualiza na mesma linha
		if sensorAtual != "" && strings.HasPrefix(msg, sensorAtual+":") {
			fmt.Printf("\r[SENSOR] %-60s", msg)
		} else {
			// resposta do sistema → nova linha e sinaliza canal
			fmt.Printf("\n[BROKER] %s\n> ", msg)
			select {
			case aguardandoResposta <- msg:
			default:
			}
		}
	}
}
