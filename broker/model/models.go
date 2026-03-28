package model

import (
	"bufio"
	"net"
	"time"
)

type Sensor struct {
	Nick string
	Tipo string
	UltimoValor string
	Addr *net.UDPAddr
	UltimoHeartBeat time.Time
	Ativo bool
	ListaInscritos []string

	
}

// model
type Atuador struct {
    Nick   string
    Tipo   string
    Ativo  bool
    Conn   net.Conn
    Reader *bufio.Reader
    Done   chan struct{} // ← sinaliza desconexão
}

type Cliente struct {
	Nick string
	Conn net.Conn
	SensoresInscritos []string
}