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

type DadosRegiao struct {
	Area            Area          `json:"area"`
	PontosDeRecarga []Localizacao `json:"pontos-de-recarga"`
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
	if erro := json.NewDecoder(file).Decode(&dadosRegiao); erro != nil {
		return DadosRegiao{}, (fmt.Errorf("Erro ao ler: %v", erro))
	}
	return dadosRegiao, nil
}
