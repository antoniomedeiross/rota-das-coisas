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
		Dispositivos[nick] = &model.Sensor{Nick: nick, Tipo: "Sensor", UltimoHeartBeat: time.Now(), Ativo: true, Addr: addr, DadosCh: make(chan string, 100)}
		log.Println("DISPOSITIVOS CONECTADOS")
		go flushSensor(nick)
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

	mutex.Lock()
	sensor, existe := Dispositivos[nickSensor]
	if !existe || sensor.Addr.String() != enviadorAddr.String() {
		mutex.Unlock()
		conn.WriteToUDP([]byte("SENSOR NÃO ENCONTRADO\n"), enviadorAddr)
		return
	}
	sensor.UltimoHeartBeat = time.Now()
	mutex.Unlock()

	//log.Println(&sensor.DadosCh)
	// joga no canal sem salvar no struct
	select {
	case sensor.DadosCh <- strings.TrimSpace(parts[1]):
	default: // canal cheio, descarta (não bloqueia)
	}

	conn.WriteToUDP([]byte("Dados recebidos\n"), enviadorAddr)
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

	if !existe {
		return "SENSOR NÃO ENCONTRADO\n"
	}
	fmt.Println(nickSensor, " - Nome do sensor")
	s := Dispositivos[nickSensor]
	s.ListaInscritos = append(s.ListaInscritos, nickClient)
	Dispositivos[nickSensor] = s

	return "ok\n"
}

func ListarDispositivosConectados(conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	i := 1
	msg := "---------- SENSORES CONECTADOS ----------\n"
	for k, a := range Dispositivos {
		ativo := "INATIVO"
		if a.Ativo {
			ativo = "ATIVO"
		}
		msg += strconv.Itoa(i) + " - " + k +
			" | TIPO: " + a.Tipo +
			" | STATUS: " + ativo +
			" | ÚLTIMA ATUALIZAÇÃO: " + a.UltimoHeartBeat.Format("02/01/06 15:04:05") +
			" | INSCRITOS: " + strconv.Itoa(len(a.ListaInscritos)) + "\n"
		i++
	}
	if i == 1 {
		msg += "Nenhum sensor conectado\n"
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

// repository.go - ComandarAtuador
func ComandarAtuador(nick string, acao string) string {
	mutex.Lock()
	a, ok := Atuadores[nick]
	mutex.Unlock()

	if !ok {
		return "ATUADOR NAO ENCONTRADO\n"
	}

	a.Mu.Lock() // ← serializa: só um comando por vez no atuador
	defer a.Mu.Unlock()

	_, err := a.Conn.Write([]byte(acao + "\n"))
	if err != nil {
		return "ERRO AO ENVIAR COMANDO\n"
	}

	resp, err := a.Reader.ReadString('\n')
	if err != nil {
		//return "ERRO AO LER RESPOSTA DO ATUADOR\n"
		select {
		case <-a.Done: // já fechado, ignora
		default:
			close(a.Done)
		}
		RemoverAtuador(nick)
		return "ATUADOR DESCONECTADO\n"

	}

	resp = strings.TrimSpace(resp)
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
    mutex.Lock()       
    defer mutex.Unlock()
    delete(Atuadores, nick)
}

func RemoverClient(nickClient string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(Clientes, nickClient)
}


func MenuHelp(conn *net.UDPConn, clientAddr *net.UDPAddr) {
    menu := "============================= HELP =============================\n" +
        "REGISTER-SENSOR <nick>     → registra o sensor\n" +
        "DATA <nick> <valor>        → envia dado do sensor\n" +
        "ESPIAR <nick>              → recebe dados do sensor em tempo real\n" +
        "PARAR-ESPIAR <nick>        → para de receber dados do sensor\n" +
        "HELP                       → exibe este menu\n" +
        "================================================================\n"
    conn.WriteToUDP([]byte(menu), clientAddr)
}

func GetAtuador(nick string) *model.Atuador {
	mutex.Lock()
	defer mutex.Unlock()
	return Atuadores[nick]
}

// //////////////////////////////////////////////////////
func flushSensor(nick string) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		mutex.Lock()
		sensor, ok := Dispositivos[nick]
		if !ok {
			mutex.Unlock()
			return
		}
		listaClientes := make([]string, len(sensor.ListaInscritos))
		copy(listaClientes, sensor.ListaInscritos)
		mutex.Unlock()

		// pega só o dado mais recente do canal, descarta o resto
		var dado string
	loop:
		for {
			select {
			case d := <-sensor.DadosCh:
				dado = d // continua drenando
			default:
				break loop // canal vazio, para
			}
		}

		if dado == "" {
			continue // nenhum dado chegou nesse intervalo
		}

		// repassa para clientes
		for _, nickCliente := range listaClientes {
			mutex.Lock()
			cliente, ok := Clientes[nickCliente]
			mutex.Unlock()
			if !ok {
				continue
			}
			go func(c net.Conn) {
				c.Write([]byte(nick + ": " + dado + "\n"))
			}(cliente.Conn)
		}
	}
}

func PararSensor(nickSensor string, nickClient string) string {
	mutex.Lock()
	defer mutex.Unlock()

	nickSensor = strings.TrimSpace(nickSensor)
	sensor, existe := Dispositivos[nickSensor]
	if !existe {
		return "SENSOR NÃO ENCONTRADO\n"
	}

	// remove o cliente da lista de inscritos
	for i, nick := range sensor.ListaInscritos {
		if nick == nickClient {
			sensor.ListaInscritos = append(sensor.ListaInscritos[:i], sensor.ListaInscritos[i+1:]...)
			return "PAROU DE SEGUIR " + nickSensor + "\n"
		}
	}

	return "VOCE NAO ESTA INSCRITO NESSE SENSOR\n"
}
