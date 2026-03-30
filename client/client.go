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
	ultimaResposta     string
	aguardandoResposta = make(chan struct{}, 1)
)

func getBrokerAddr() string {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		return BROKER_ADDR
	}
	if strings.Contains(addr, ":") {
		return addr
	}
	return addr + ":9091"
}

func main() {
	log.Println(getBrokerAddr())

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

	enviar("REGISTER-CLIENT " + nick)
	resp := receber()
	if !strings.Contains(resp, "CLIENTE CONECTADO") {
		log.Fatalln("Erro ao registrar:", resp)
	}

	fmt.Println("\nConectado ao broker como:", nick)

	go escutarBroker()

	for {
		mostrarMenu()
		fmt.Print("> ")
		scanner.Scan()
		opcao := strings.TrimSpace(scanner.Text())

		switch opcao {
		case "1":
			listarSensores()
		case "2":
			seguirSensor(scanner)
		case "3":
			pararSensor()
		case "4":
			listarAtuadores()
		case "5":
			comandarAtuador(scanner)
			// no switch do main
		case "6":
			listarClientes() // ← adiciona
		case "7": // ← era 6
			enviar("QUIT")
			fmt.Println("Saindo...")
			os.Exit(0)
		default:
			fmt.Println("Opção inválida!")
		}
	}
}

func mostrarMenu() {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("         BROKER IoT - CLIENTE")
	fmt.Println("========================================")
	if sensorAtual != "" {
		fmt.Println("  [ouvindo sensor: " + sensorAtual + "]")
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

func listarSensores() {
	enviar("LIST-SENSORES")
	<-aguardandoResposta
}

func listarAtuadores() {
	enviar("LIST-ATUADORES")
	<-aguardandoResposta
}

// adiciona a função
func listarClientes() {
    enviar("LIST-CLIENTES")
    <-aguardandoResposta
}

func seguirSensor(scanner *bufio.Scanner) {
	if sensorAtual != "" {
		fmt.Println("Você já está seguindo o sensor:", sensorAtual)
		fmt.Println("Pare de seguir antes de escolher outro (opção 3)")
		return
	}
	fmt.Print("Nick do sensor: ")
	scanner.Scan()
	nickSensor := strings.TrimSpace(scanner.Text())
	enviar("SEGUIR-SENSOR " + nickSensor)
	<-aguardandoResposta

	// só seta sensorAtual se o broker confirmou
	if strings.Contains(ultimaResposta, "ok") {
		sensorAtual = nickSensor
		fmt.Println("Seguindo sensor:", nickSensor, "(os dados aparecerão automaticamente)")
	}
}

func pararSensor() {
	if sensorAtual == "" {
		fmt.Println("Você não está seguindo nenhum sensor.")
		return
	}
	enviar("PARAR-SENSOR " + sensorAtual)
	<-aguardandoResposta
	sensorAtual = ""
}

func comandarAtuador(scanner *bufio.Scanner) {
	fmt.Print("Nick do atuador: ")
	scanner.Scan()
	nickAtuador := strings.TrimSpace(scanner.Text())

	fmt.Print("Ação (ON/OFF): ")
	scanner.Scan()
	acao := strings.TrimSpace(scanner.Text())

	enviar("COMMAND " + nickAtuador + " " + acao)
	<-aguardandoResposta
}

func enviar(msg string) {
	conn.Write([]byte(msg + "\n"))
}

func receber() string {
	resp, err := leitor.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(resp)
}

func escutarBroker() {
	var ultimaEraSensor bool
	for {
		msg, err := leitor.ReadString('\n')
		if err != nil {
			fmt.Println("\nConexão com o broker encerrada.")
			os.Exit(0)
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}

		// verifica se é dado do sensor que estamos seguindo
		ehSensor := sensorAtual != "" && strings.HasPrefix(msg, sensorAtual+":")

		if ehSensor && ultimaEraSensor {
			// apaga linha anterior do sensor
			fmt.Print("\033[1A\033[2K")
		}

		ultimaResposta = msg
		fmt.Println("\n[broker] " + msg)
		ultimaEraSensor = ehSensor

		// sinaliza que chegou uma resposta
		select {
		case aguardandoResposta <- struct{}{}:
		default:
		}
	}
}
