package dataJson

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type Mensagem struct {
	Tipo     string `json:"tipo"`
	Conteudo string `json:"conteudo"`
	Origem   string `json:"origem"`
}

type Area struct {
	Latitude_max  float64 `json:"latitude_max"`
	Latitude_min  float64 `json:"latitude_min"`
	Longitude_min float64 `json:"longitude_min"`
	Longitude_max float64 `json:"longitude_max"`
}

type Localizacao struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Ponto struct {
	ID        int     `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DadosRegiao struct {
	Area            Area    `json:"area_cobertura"`
	PontosDeRecarga []Ponto `json:"pontos-de-recarga"`
}

func ReceiveMessage(conexao net.Conn) (Mensagem, error) {
	var msg Mensagem
	decoder := json.NewDecoder(conexao)
	erro := decoder.Decode(&msg)
	if erro != nil {
		return msg, fmt.Errorf("erro: %v", erro)
	}
	return msg, nil
}

func SendMessage(conexao net.Conn, msg Mensagem) error {
	encoder := json.NewEncoder(conexao)
	erro := encoder.Encode(msg)
	if erro != nil {
		return fmt.Errorf("erro: %v", erro)
	}
	return nil
}

func OpenFile(arquivo string) (DadosRegiao, error) {
	path := filepath.Join("internal", "dataJson", arquivo)
	file, erro := os.Open(path)
	if erro != nil {
		return DadosRegiao{}, (fmt.Errorf("Erro ao abrir: %v", erro))
	}
	defer file.Close()

	var dadosRegiao DadosRegiao
	erro = json.NewDecoder(file).Decode(&dadosRegiao)
	if erro != nil {
		return DadosRegiao{}, (fmt.Errorf("Erro ao ler: %v", erro))
	}
	return dadosRegiao, nil
}

func GetPontosDeRecargaJson() ([]Ponto, error) {
	dadosRegiao, erro := OpenFile("regiao.json")

	if erro != nil {
		return dadosRegiao.PontosDeRecarga, fmt.Errorf("Erro ao carregar dados JSON")
	}
	return dadosRegiao.PontosDeRecarga, nil
}

func GetTotalPontosJson() int {
	pontos, erro := GetPontosDeRecargaJson()
	if erro != nil {
		return -1
	}
	return len(pontos)
}

func GetPontoId(id int) (Ponto, int) {
	dadosRegiao, erro := OpenFile("regiao.json")
	if erro != nil {
		return Ponto{}, 1 //Erro ao carregar dados JSON
	}

	for _, ponto := range dadosRegiao.PontosDeRecarga {
		if ponto.ID == id {
			return ponto, 0
		}
	}

	return Ponto{}, 2 //Erro ao localizar ponto
}
