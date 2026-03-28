package repository

import (
	"broker/model"
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

// map[nick] → struct sensor
var Dispositivos = make(map[string]*model.Sensor)

// map[nick] → struct atuador
var Atuadores = make(map[string]*model.Atuador)

// map[nick] → struct clients
var Clientes = make(map[string]*model.Cliente)

func SalvarSensor(nick string, addr *net.UDPAddr) string {
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	for i := range Dispositivos {
		if i == nick {
			log.Println("NICK DE DISPOSITIVO JÁ EXISTENTE")
			return "NICK DE DISPOSITIVO JÁ EXISTENTE \n"
		}
	}

	if len(nick) > 0 {
		nick = strings.TrimSpace(nick)
		Dispositivos[nick] = &model.Sensor{Nick: nick, Tipo: "Sensor", UltimoValor: "sensor desligado", UltimoHeartBeat: time.Now(), Ativo: true, Addr: addr}
		log.Println("DISPOSITIVOS CONECTADOS")
		return "DISPOSITIVO CONECTADO \n"
	}
	return "ERRO AO CONECTAR SENSOR"
}

func SalvarAtuador(nick string, conn net.Conn, reader *bufio.Reader) string {
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	for i := range Atuadores {
		if i == nick {
			log.Println("NICK DE DISPOSITIVO JÁ EXISTENTE")
			return "NICK DE DISPOSITIVO JÁ EXISTENTE \n"
		}
	}

	if len(nick) > 0 {
		nick = strings.TrimSpace(nick)
		Atuadores[nick] = &model.Atuador{Nick: nick, Tipo: "Alarme", Ativo: false, Conn: conn, Reader: reader, Done: make(chan struct{})}
		//log.Println("ATUADOR CONECTADO")
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

	if len(nick) > 0 {
		nick = strings.TrimSpace(nick)
		Clientes[nick] = &model.Cliente{Nick: nick, Conn: conn}
		log.Println("CLIENTE CONECTADO")
		return "CLIENTE CONECTADO \n"
	}
	return "ERRO AO CONECTAR CLIENTE"
}

func EnviarDados(nickValue string, conn *net.UDPConn, enviadorAddr *net.UDPAddr) {

	parts := strings.SplitN(string(nickValue), " ", 2)
	nickSensor := strings.TrimSpace(parts[0])

	// pega dados protegidos
	mutex.Lock()

	sensor, existe := Dispositivos[nickSensor]
	if !existe || sensor.Addr.String() != enviadorAddr.String() {
		mutex.Unlock()
		conn.WriteToUDP([]byte("SENSOR NÃO ENCONTRADO\n"), enviadorAddr)
		return
	}

	// atualiza heartbeat
	sensor.UltimoHeartBeat = time.Now()
	Dispositivos[nickSensor] = sensor

	// copia lista (IMPORTANTÍSSIMO)
	listaClientes := make([]string, len(sensor.ListaInscritos))
	copy(listaClientes, sensor.ListaInscritos)

	mutex.Unlock() // libera antes de I/O

	// responde sensor
	conn.WriteToUDP([]byte("Dados recebidos com sucesso"), enviadorAddr)

	// envia para clientes
	for _, nickCliente := range listaClientes {
		cliente, ok := Clientes[nickCliente]
		if !ok {
			continue
		}

		go func(c net.Conn) { //EVIA VIA CONCORRENCIA
			_, err := c.Write([]byte(parts[1]))
			if err != nil {
				log.Println("Erro ao enviar para cliente:", err)
			}
		}(cliente.Conn)
	}
}

// ///////////////////////
// ////////////////////////
// //////////////////////
// MUDAR LOGICA PARA TCPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP
func SeguirSensor(nickSensor string, nickClient string) string {
	// TRAVA A ESTRUTURA
	mutex.Lock()
	defer mutex.Unlock()

	nickSensor = strings.TrimSpace(nickSensor)

	_, existe := Dispositivos[nickSensor]

	fmt.Println(existe)
	if existe {
		fmt.Println(nickSensor, " - Nome do sensor")
		s := Dispositivos[nickSensor]
		s.ListaInscritos = append(s.ListaInscritos, nickClient)
		Dispositivos[nickSensor] = s

	} else {
		fmt.Println("SENSOR INVÁLIDO!")
	}

	return "ok"
}

func ListarDispositivosConectados(conn net.Conn) {
	i := 1
	msg := "---------- SENSORES CONECTADOS ----------\n"

	for k, a := range Dispositivos {
		msg += strconv.Itoa(i) + " - " + k +
			" | TIPO: " + a.Tipo +
			" | ATIVO: " + strconv.FormatBool(a.Ativo) +
			" | ÚLTIMA ATUALIZAÇÃO: " + a.UltimoHeartBeat.Format("02/01/06 15:04:01") + "\n"
		i++
	}

	fmt.Println(msg)
	conn.Write([]byte(msg))
}

func ListarAtuadoresConectados(conn net.Conn) {
    mutex.Lock()
    defer mutex.Unlock()

    i := 1
    msg := "---------- ATUADORES CONECTADOS ----------\n"
    for k, a := range Atuadores {
        ativo := "DESLIGADO"
        if a.Ativo {
            ativo = "LIGADO"
        }
        msg += strconv.Itoa(i) + " - " + k +
            " | TIPO: " + a.Tipo +
            " | STATUS: " + ativo + "\n"
        i++
    }
    if i == 1 {
        msg += "Nenhum atuador conectado\n"
    }
    fmt.Println(msg)
    conn.Write([]byte(msg))
}

func ListarClientesConectados(conn net.Conn) {
	i := 1
	var msg string = "---------- CLIENTES CONECTADOS ---------- \n"
	for k := range Clientes {
		msg += strconv.Itoa(i) + " - " + k + "\n"
		i++
	}
	fmt.Println(msg)
	conn.Write([]byte(msg))
}

func ComandarAtuador(nick string, acao string) string {
	mutex.Lock()
	a, ok := Atuadores[nick]
	mutex.Unlock()

	if !ok {
		return "ATUADOR NAO ENCONTRADO\n"
	}

	msg := acao + "\n"
	fmt.Println(acao + "-------------------------------------------")

	_, err := a.Conn.Write([]byte(msg))
	if err != nil {
		return "ERRO AO ENVIAR COMANDO\n"
	}
	fmt.Println("DEBUG-1-------------------------------------------")

	reader := a.Reader
	fmt.Println("DEBUG-2-------------------------------------------")

	resp, err := reader.ReadString('\n')
	fmt.Println("DEBUG-3-------------------------------------------")

	if err != nil {
		return "ERRO AO LER RESPOSTA DO ATUADOR\n"
	}

	fmt.Println("DEBUG-4-------------------------------------------")

	fmt.Println("aqui-" + resp + "-aqui")
	resp = strings.TrimSpace(resp)
	fmt.Println("aqui-" + resp + "-aqui")

	fmt.Println("DEBUG-5-------------------------------------------")

	switch resp {
	case "ATUADOR LIGADO":
		a.Ativo = true
	case "ATUADOR DESLIGADO":
		a.Ativo = false
	case "ATUADOR JA ESTA DESLIGADO":
		a.Ativo = false
	case "ATUADOR JA ESTA LIGADO":
		a.Ativo = true
	default:
		return "O ATUADOR ENVIOU UMA RESPOSTA INVALIDA\n"
	}

	fmt.Println("DEBUG-6-------------------------------------------")

	return resp + "\n"
}

// NAO TO USANDO MAIIIS
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

func RemoverAtuador(nick string) {
	for a := range Atuadores {
		if a == nick {
			delete(Atuadores, nick)
		}
	}
}

func MenuHelp(conn *net.UDPConn, clientAddr *net.UDPAddr) {
	menu := ("============================= HELP INCOMPLETO ============================= \n" +
		"REGISTER nickSensor --> registra o sensor \n")

	conn.WriteToUDP([]byte(menu), clientAddr)
}


func GetAtuador(nick string) *model.Atuador {
    mutex.Lock()
    defer mutex.Unlock()
    return Atuadores[nick]
}