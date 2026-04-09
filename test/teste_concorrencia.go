package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const BROKER_ADDR = "192.168.0.103:9091"

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

func simularCliente(nick string, atuador string, acoes []string, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", getBrokerAddr())
	if err != nil {
		fmt.Printf("[%s] ERRO ao conectar: %v\n", nick, err)
		return
	}
	defer conn.Close()

	leitor := bufio.NewReader(conn)

	// registra
	conn.Write([]byte("REGISTER-CLIENT " + nick + "\n"))
	resp, _ := leitor.ReadString('\n')
	if !strings.Contains(resp, "CLIENTE CONECTADO") {
		fmt.Printf("[%s] ERRO ao registrar: %s\n", nick, strings.TrimSpace(resp))
		return
	}
	fmt.Printf("[%s] conectado!\n", nick)

	// envia todos os comandos e lê uma resposta por comando
	for _, acao := range acoes {
		conn.Write([]byte("COMMAND " + atuador + " " + acao + "\n"))
		inicio := time.Now()

		resp, err := leitor.ReadString('\n')
		duracao := time.Since(inicio)

		if err != nil {
			fmt.Printf("[%s] COMMAND %s → ERRO DE LEITURA\n", nick, acao)
			continue
		}

		fmt.Printf("[%s] COMMAND %-4s → %-40s (%.0fms)\n",
			nick, acao, strings.TrimSpace(resp), float64(duracao.Milliseconds()))
	}
}

func main() {
	atuador := os.Getenv("ATUADOR_NICK") //
	if len(os.Args) > 1 {
		atuador = os.Args[1]
	}

	fmt.Println("================================================")
	fmt.Println("     TESTE DE CONCORRÊNCIA - BROKER IoT")
	fmt.Println("================================================")
	fmt.Println("Atuador alvo :", atuador)
	fmt.Println("Broker       :", getBrokerAddr())
	fmt.Println("================================================")
	fmt.Println()

	// os dois clientes enviam a mesma sequência de ações
	acoes := []string{"ON", "ON", "OFF", "OFF", "OFF"}

	var wg sync.WaitGroup
	wg.Add(2)

	inicio := time.Now()

	// dispara os dois clientes simultaneamente
	go simularCliente("cliente-1", atuador, acoes, &wg)
	go simularCliente("cliente-2", atuador, acoes, &wg)

	wg.Wait()

	fmt.Println()
	fmt.Printf("Duração total: %.0fms\n", float64(time.Since(inicio).Milliseconds()))
	fmt.Println("================================================")
}