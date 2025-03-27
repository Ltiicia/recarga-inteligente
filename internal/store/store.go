package store

import (
	"net"
	"sync"
)

type ConnectionStore struct {
	mutex           sync.Mutex
	veiculos        map[net.Conn]string
	pontosDeRecarga map[net.Conn]string
}

func NewConnectionStore() *ConnectionStore {
	return &ConnectionStore{
		veiculos:        make(map[net.Conn]string),
		pontosDeRecarga: make(map[net.Conn]string),
	}
}

func (connection *ConnectionStore) AddConnection(conexao net.Conn, clientType string) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	if clientType == "veiculo" {
		connection.veiculos[conexao] = conexao.RemoteAddr().String()
	} else if clientType == "ponto-de-recarga" {
		connection.pontosDeRecarga[conexao] = conexao.RemoteAddr().String()
	}
}

func (connection *ConnectionStore) RemoveConnection(conexao net.Conn) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	delete(connection.veiculos, conexao)
	delete(connection.pontosDeRecarga, conexao)
	conexao.Close()
}
