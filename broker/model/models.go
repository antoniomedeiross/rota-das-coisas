package model

import (
	"net"
	"sync"
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
type Atuador struct {
	Nick string
	Tipo string
	Ativo bool
	Conn net.Conn
	Mu   sync.Mutex
}

type Cliente struct {
	Nick string
	Conn net.Conn
	SensoresInscritos []string
}