package model

import (
	"net"
)

type Sensor struct {
	Nick string
	Tipo string
	UltimoValor string
	Addr *net.UDPAddr
	ListaInscritos []*net.UDPAddr
}

type Atuador struct {
	Nick string
	Tipo string
	Ativo bool
	Conn net.Conn
	//Mu   sync.Mutex
}

type Cliente struct {
	Nick string
	Conn net.Conn
	SensoresInscritos []string
}