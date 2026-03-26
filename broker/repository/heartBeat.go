package repository

import (
	"log"
	"time"
)

func HeartBeat() {
	// preciso iterarar sobre os sensores ver quandoe eles mandaram a ultima msg e ver se faz menos de 60 segundos, se sim tira eles da lista
	for {
		mutex.Lock()
		for nick, s := range Dispositivos {
			if time.Since(s.UltimoHeartBeat) > 60*time.Second {
				s.Ativo = false
				Dispositivos[nick] = s
			}

			if time.Since(s.UltimoHeartBeat) > 120*time.Second {
				delete(Dispositivos, nick)
				log.Println("Sensor removido:", nick)
			}
		}
		mutex.Unlock()
	}
}
