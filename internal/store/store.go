package store

import (
	"fmt"
	"net"
	"recarga-inteligente/internal/dataJson"
	"sort"
	"sync"
)

type ConnectionStore struct {
	mutex                 sync.Mutex
	veiculos              map[net.Conn]string
	pontosDeRecarga       map[net.Conn]int
	idsCadastrados        []int
	filasDosPontos        map[int][]dataJson.Veiculo
	disponibilidadePontos map[int]bool
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

		idsCadastrados: idsJson,

		filasDosPontos:        make(map[int][]dataJson.Veiculo),
		disponibilidadePontos: make(map[int]bool),
	}
}

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
	}
	fmt.Printf("Placa removida da conexão: %s\n", connection.veiculos[conexao])
	delete(connection.veiculos, conexao)

	conexao.Close()
}

// Retorna um mapa de todas as conexões de pontos de recarga
func (connection *ConnectionStore) GetPontosMap() map[net.Conn]int {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

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

func (connection *ConnectionStore) GetVeiculoPlaca(conexao net.Conn) string {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	return connection.veiculos[conexao]
}

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

func (connection *ConnectionStore) GetTodasPlacasAtivas() []string {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	placas := make([]string, 0, len(connection.veiculos))
	for _, placa := range connection.veiculos {
		placas = append(placas, placa)
	}

	return placas
}

func (connection *ConnectionStore) GetFilaPorPonto(pontoID int) []dataJson.Veiculo {
	return connection.filasDosPontos[pontoID]
}

func (connection *ConnectionStore) AtualizarFilaDoPonto(pontoID int, fila []dataJson.Veiculo) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	connection.filasDosPontos[pontoID] = fila
}

func (connection *ConnectionStore) AdicionarVeiculoNaFila(pontoID int, veiculo dataJson.Veiculo) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	fila := connection.filasDosPontos[pontoID]
	connection.filasDosPontos[pontoID] = append(fila, veiculo)
}

func (connection *ConnectionStore) RemoverVeiculoDaFila(pontoID int, placa string) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	fila := connection.filasDosPontos[pontoID]
	novaFila := []dataJson.Veiculo{}
	for _, veiculo := range fila {
		if veiculo.Placa != placa {
			novaFila = append(novaFila, veiculo)
		}
	}
	connection.filasDosPontos[pontoID] = novaFila
}

func (c *ConnectionStore) VeiculoEstaEmFila(placa string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, fila := range c.filasDosPontos {
		for _, veiculo := range fila {
			if veiculo.Placa == placa {
				return true
			}
		}
	}
	return false
}

func (connection *ConnectionStore) PlacaJaEmUso(placa string, conexaoAtual net.Conn) bool {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	for conn, p := range connection.veiculos {
		if p == placa && conn != conexaoAtual {
			return true
		}
	}
	return false
}

func (connection *ConnectionStore) RemoverPlacaAtiva(conexao net.Conn) {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	for c, _ := range connection.veiculos {
		if c == conexao {
			connection.veiculos[c] = ""
			delete(connection.veiculos, c)
			connection.RemoveConnection(c)
		}
	}
	if _, existe := connection.veiculos[conexao]; existe {
		delete(connection.veiculos, conexao)
	}
	connection.veiculos[conexao] = ""
}
