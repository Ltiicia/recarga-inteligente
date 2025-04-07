package store

import (
	"net"
	"recarga-inteligente/internal/dataJson"
	"sort"
	"sync"
)

type ConnectionStore struct {
	mutex           sync.Mutex
	veiculos        map[net.Conn]string // Agora armazena a placa em vez do endereço
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

// Modifique a função para garantir que a placa seja usada corretamente:
func (connection *ConnectionStore) AddVeiculo(conexao net.Conn, placa string) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	if placa == "" {
		// Se não tiver placa, usa o endereço como identificação temporária
		connection.veiculos[conexao] = conexao.RemoteAddr().String()
	} else {
		// Se tiver placa, usa a placa como identificação
		connection.veiculos[conexao] = placa
	}
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

// Retorna um mapa de todas as conexões de pontos de recarga
func (connection *ConnectionStore) GetPontosMap() map[net.Conn]int {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	// Crie uma cópia para evitar race conditions
	pontosCopy := make(map[net.Conn]int)
	for conn, id := range connection.pontosDeRecarga {
		pontosCopy[conn] = id
	}

	return pontosCopy
}

// Retorna a conexão de um ponto pelo seu ID
func (connection *ConnectionStore) GetConexaoPorID(id int) net.Conn {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	for conn, pontoID := range connection.pontosDeRecarga {
		if pontoID == id {
			return conn
		}
	}

	return nil
}

// Adicionar função para obter a placa do veículo:
func (connection *ConnectionStore) GetVeiculoPlaca(conexao net.Conn) string {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	return connection.veiculos[conexao]
}

// Retorna a conexão de um veículo pela sua placa
func (connection *ConnectionStore) GetConexaoPorPlaca(placa string) net.Conn {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	for conn, placaConn := range connection.veiculos {
		if placaConn == placa {
			return conn
		}
	}

	return nil
}

// GetTodasPlacasAtivas returns all active vehicle plates
func (connection *ConnectionStore) GetTodasPlacasAtivas() []string {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	placas := make([]string, 0, len(connection.veiculos))
	for _, placa := range connection.veiculos {
		placas = append(placas, placa)
	}

	return placas
}
