package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

func getServerAddr() string {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "localhost" // fallback
	}
	fmt.Println(addr)

	return addr
}

func main() {
	// Configura o rand
	nick, err := os.Hostname()

	if err != nil {
		log.Fatalf("Erro ao pegar nick do host")
	}

	//ladrr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}    // porta dinâmica
	radrr := &net.UDPAddr{IP: net.ParseIP(getServerAddr()), Port: 9090} // servidor

	conn, err := net.DialUDP("udp", nil, radrr)
	// conn, err := net.DialUDP("udp", ladrr, radrr)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close()

	fmt.Println("Cliente UDP iniciado. Conectado ao servidor no adr 9090")

	// 1. Primeiro envia REGISTER

	_, err = conn.Write([]byte("REGISTER-SENSOR " + nick))
	if err != nil {
		log.Fatalf("Erro ao registrar: %v", err)
	}
	fmt.Println("REGISTER-SENSOR " + nick + " enviado")

	// Goroutine para receber respostas do servidor
	go func() {
		buffer := make([]byte, 1024)
		for {
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			n, err := conn.Read(buffer)
			if err != nil {
				// Timeout ou erro, continua
				continue
			}

			resposta := string(buffer[:n])
			fmt.Printf("Resposta do servidor: %s\n", resposta)
		}
	}()

	// 2. Loop para enviar DATA a cada 0.1 segundos
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Gera temperatura aleatória
		temp := strconv.Itoa(rand.Intn(100))

		// Envia DATA
		mensagem := []byte("DATA " + nick + " " + temp)
		_, err := conn.Write(mensagem)

		if err != nil {
			log.Printf("Erro ao enviar DATA: %v", err)
		} else {
			fmt.Printf("Enviado: DATA %s %s\n", nick, temp)
		}
	}
}
