package store

import (
	"net"
	"recarga-inteligente/internal/dataJson"
	"sync"
)

type ConnectionStore struct {
	mutex           sync.Mutex
	veiculos        map[net.Conn]string
	pontosDeRecarga map[net.Conn]int
}

var idAtual = 1

func NewConnectionStore() *ConnectionStore {
	return &ConnectionStore{
		veiculos:        make(map[net.Conn]string),
		pontosDeRecarga: make(map[net.Conn]int),
	}
}

func (connection *ConnectionStore) AddVeiculo(conexao net.Conn) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	connection.veiculos[conexao] = conexao.RemoteAddr().String()
}

func (connection *ConnectionStore) AddPontoRecarga(conexao net.Conn) int {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	if idAtual > dataJson.GetTotalPontosJson() {
		return -1
	}
	id := idAtual
	connection.pontosDeRecarga[conexao] = id
	idAtual++
	return id
}

func (connection *ConnectionStore) GetIdPonto(conexao net.Conn) int {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	id := connection.pontosDeRecarga[conexao]
	return id
}

func (connection *ConnectionStore) GetTotalPontosConectados() int {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	return len(connection.pontosDeRecarga)
}

func (connection *ConnectionStore) RemoveConnection(conexao net.Conn) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	delete(connection.veiculos, conexao)
	delete(connection.pontosDeRecarga, conexao)
	conexao.Close()
}
