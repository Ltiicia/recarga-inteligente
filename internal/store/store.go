package store

import (
	"net"
	"recarga-inteligente/internal/dataJson"
	"sort"
	"sync"
)

type ConnectionStore struct {
	mutex           sync.Mutex
	veiculos        map[net.Conn]string
	pontosDeRecarga map[net.Conn]int
	idsCadastrados  []int
}

func NewConnectionStore() *ConnectionStore {
	total := dataJson.GetTotalPontosJson()
	idsJson := make([]int, 0, total)
	for i := 1; i <= total; i++ {
		idsJson = append(idsJson, i)
	}

	return &ConnectionStore{
		veiculos:        make(map[net.Conn]string),
		pontosDeRecarga: make(map[net.Conn]int),
		idsCadastrados:  idsJson,
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
	if len(connection.idsCadastrados) == 0 {
		return -1
	}
	id := connection.idsCadastrados[0]
	connection.idsCadastrados = connection.idsCadastrados[1:] //remove o id utilizado
	connection.pontosDeRecarga[conexao] = id
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

	id, idExiste := connection.pontosDeRecarga[conexao]
	if idExiste {
		connection.idsCadastrados = append(connection.idsCadastrados, id) //retorna o id para a lista
		sort.Ints(connection.idsCadastrados)                              //ordena
		delete(connection.pontosDeRecarga, conexao)
	} else {
		delete(connection.veiculos, conexao)
	}
	conexao.Close()
}
