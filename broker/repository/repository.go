package repository

import (
	"broker/model"
	"sync"
	"strings"
	"log"
	"fmt"
	"net"
	"strconv"
	"bufio"
)


var mutex sync.Mutex

// map[nick] → struct sensor
var Dispositivos = make(map[string]model.Sensor)

// map[nick] → struct atuador
var Atuadores = make(map[string]model.Atuador)

// map[nick] → struct clients
var Clientes = make(map[string]model.Cliente)



func SalvarSensor(nick string, addr *net.UDPAddr) string{
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	for i := range Dispositivos {
		if i == nick {
			log.Println("NICK DE DISPOSITIVO JÁ EXISTENTE")
			return "NICK DE DISPOSITIVO JÁ EXISTENTE \n"
		}
	}

	if len(nick) > 0	{
		nick = strings.TrimSpace(nick)
		Dispositivos[nick] = model.Sensor{Nick: nick, Tipo: "Sensor", UltimoValor: "sensor desligado", Addr: addr}
		log.Println("DISPOSITIVOS CONECTADOS")
		return "DISPOSITIVO CONECTADO \n"
	}
	return "ERRO AO CONECTAR SENSOR"
}

func SalvarAtuador(nick string, conn net.Conn) string {
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	for i := range Atuadores {
		if i == nick {
			log.Println("NICK DE DISPOSITIVO JÁ EXISTENTE")
			return "NICK DE DISPOSITIVO JÁ EXISTENTE \n"
		}
	}

	if len(nick) > 0	{
		nick = strings.TrimSpace(nick)
		Atuadores[nick] = model.Atuador{Nick: nick, Tipo: "Alarme", Ativo: false, Conn: conn}
		log.Println("ATUADOR CONECTADO")
		return "ATUADOR CONECTADO \n"
	}
	return "ERRO AO CONECTAR ATUADOR"
}

func SalvarCliente(nick string, conn net.Conn) string {
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	for i := range Clientes {
		if i == nick {
			log.Println("NICK DE CLIENTE JÁ EXISTENTE")
			return "NICK DE CLIENTE JÁ EXISTENTE \n"
		}
	}

	if len(nick) > 0	{
		nick = strings.TrimSpace(nick)
		Clientes[nick] = model.Cliente{Nick: nick, Conn: conn}
		log.Println("CLIENTE CONECTADO")
		return "CLIENTE CONECTADO \n"
	}
	return "ERRO AO CONECTAR CLIENTE"
}

func EnviarDados(nickValue string, conn *net.UDPConn, enviadorAddr *net.UDPAddr) {
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	parts := strings.SplitN(string(nickValue), " ", 2)
	//nick := parts[0]
	nickSensor := strings.TrimSpace(parts[0])
	// parts[1] = valor
	if Dispositivos[nickSensor].Addr.String() == enviadorAddr.String() {
		listaAddrs := Dispositivos[nickSensor].ListaInscritos
		
		//retorna para o sensor se os dados foram recebidos
		conn.WriteToUDP([]byte("Dados recebidos com sucesso"), enviadorAddr)
		
		for _, a := range listaAddrs {
			conn.WriteToUDP([]byte(parts[1]+"\n"), a)
		}
		return
	}

	conn.WriteToUDP([]byte("SEU NICK NÃO VÁLIDO \n"), enviadorAddr)

}

func SeguirSensor(nickSensor string, appAddr *net.UDPAddr) {
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	nickSensor = strings.TrimSpace(nickSensor)

	_, existe := Dispositivos[nickSensor]

	fmt.Println(existe)
	if existe {
		fmt.Println(nickSensor, " - Nome do sensor")
		s := Dispositivos[nickSensor]
		s.ListaInscritos = append(s.ListaInscritos, appAddr)
		Dispositivos[nickSensor] = s
	
	} else {
		fmt.Println("SENSOR INVÁLIDO!")
	}
}

func ListarDispositivosConectados(conn *net.UDPConn, clientAddr *net.UDPAddr) {
	i := 1
	var msg string = "---------- SENSORES CONECTADOS ---------- \n"
	for k := range Dispositivos {
		msg += strconv.Itoa(i)+" - "+k+"\n"; 
		i++
	}
	fmt.Println(msg)
	conn.WriteToUDP([]byte(msg), clientAddr);
	
}


func ComandarAtuador(nick string, acao string) string {
	mutex.Lock()
	a, ok := Atuadores[nick]
	defer mutex.Unlock()

	
	
	if !ok {
		return "ATUADOR NAO ENCONTRADO\n"
	}
	// envia comando
	msg := acao + "\n"
	fmt.Println(acao)
	a.Conn.Write([]byte(msg))


	reader := bufio.NewReader(a.Conn)

	resp, err := reader.ReadString('\n')
	if err != nil {
		return "ERRO AO LER RESPOSTA DO ATUADOR\n"
	}

	return resp
}

// repository/atuador.go
func SetAtuador(nick string, ativo bool) bool {
	mutex.Lock()
	defer mutex.Unlock()

	a, ok := Atuadores[nick]
	if !ok {
		return false
	}

	a.Ativo = ativo
	Atuadores[nick] = a
	return true
}

func MenuHelp(conn *net.UDPConn, clientAddr *net.UDPAddr) {
	menu := ("============================= HELP INCOMPLETO ============================= \n" +
		"REGISTER nickSensor --> registra o sensor \n")

	conn.WriteToUDP([]byte(menu), clientAddr)
}

