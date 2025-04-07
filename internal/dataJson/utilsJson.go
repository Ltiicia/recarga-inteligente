package dataJson

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Mensagem struct {
	Tipo     string `json:"tipo"`
	Conteudo string `json:"conteudo"`
	Origem   string `json:"origem"`
}

type DadosJson struct {
	Titulo string      `json:"titulo"`
	Dados  DadosRegiao `json:"dados"`
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
	Area            Area    `json:"area-cobertura"`
	PontosDeRecarga []Ponto `json:"pontos-de-recarga"`
}

type Veiculo struct {
	Placa    string    `json:"placa"`
	Recargas []Recarga `json:"recargas,omitempty"`
}

type Recarga struct {
	Data    string  `json:"data"`
	PontoID int     `json:"ponto_id"`
	Valor   float64 `json:"valor"`
}

type DadosVeiculos struct {
	Veiculos []Veiculo `json:"veiculos"`
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

func ReceiveDadosJson(conexao net.Conn) (DadosJson, error) {
	var payload DadosJson
	decoder := json.NewDecoder(conexao)
	erro := decoder.Decode(&payload)
	if erro != nil {
		return DadosJson{}, fmt.Errorf("Erro ao receber dados JSON: %v", erro)
	}
	return payload, nil
}

func SendDadosJson(conexao net.Conn, titulo string, dados DadosRegiao) error {
	payload := DadosJson{
		Titulo: titulo,
		Dados:  dados,
	}

	encoder := json.NewEncoder(conexao)
	erro := encoder.Encode(payload)
	if erro != nil {
		return fmt.Errorf("Erro ao enviar dados JSON: %v", erro)
	}
	return nil
}

func OpenFile(arquivo string) (DadosRegiao, error) {
	path := filepath.Join("app", "internal", "dataJson", arquivo)
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

// Cliente
func ReceiveDadosRegiao(conexao net.Conn) (DadosRegiao, string, error) {
	dados, erro := ReceiveDadosJson(conexao)
	if erro != nil {
		return DadosRegiao{}, dados.Titulo, erro
	}
	if dados.Titulo != "dados-regiao" {
		return DadosRegiao{}, dados.Titulo, fmt.Errorf("tipo de dados inesperado: %s", dados.Titulo)
	}
	return dados.Dados, dados.Titulo, nil
}

// Servidor
func SendDadosRegiao(conexao net.Conn) error {
	dadosRegiao, erro := OpenFile("regiao.json")
	if erro != nil {
		return fmt.Errorf("Erro ao carregar dados da regiao do JSON: %v", erro)
	}
	return SendDadosJson(conexao, "dados-regiao", dadosRegiao)
}

func GetTotalPontosJson() int {
	pontos, erro := GetPontosDeRecargaJson()
	if erro != nil {
		return -1
	}
	return len(pontos)
}

func GetPontosDeRecargaJson() ([]Ponto, error) {
	dadosRegiao, erro := OpenFile("regiao.json")
	if erro != nil {
		return dadosRegiao.PontosDeRecarga, fmt.Errorf("Erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.PontosDeRecarga, nil
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

func SalvarVeiculo(placa string) error {
	path := filepath.Join("app", "internal", "dataJson", "veiculos.json")

	// Ler arquivo existente
	file, err := os.Open(path)
	if err != nil {
		// Se o arquivo não existe, cria um novo
		if os.IsNotExist(err) {
			dadosVeiculos := DadosVeiculos{
				Veiculos: []Veiculo{
					{Placa: placa, Recargas: []Recarga{}},
				},
			}
			return salvarDadosVeiculos(path, dadosVeiculos)
		}
		return err
	}
	defer file.Close()

	// Decodificar dados existentes
	var dadosVeiculos DadosVeiculos
	if err := json.NewDecoder(file).Decode(&dadosVeiculos); err != nil {
		return err
	}

	// Verificar se o veículo já existe
	for _, v := range dadosVeiculos.Veiculos {
		if v.Placa == placa {
			return nil // Veículo já cadastrado
		}
	}

	// Adicionar novo veículo
	dadosVeiculos.Veiculos = append(dadosVeiculos.Veiculos, Veiculo{
		Placa:    placa,
		Recargas: []Recarga{},
	})

	return salvarDadosVeiculos(path, dadosVeiculos)
}

func RegistrarRecarga(placa string, pontoID int, valor float64) error {
	// Verificar entrada
	if placa == "" {
		return fmt.Errorf("placa do veículo não pode ser vazia")
	}
	if pontoID <= 0 {
		return fmt.Errorf("ID do ponto de recarga inválido: %d", pontoID)
	}
	if valor <= 0 {
		return fmt.Errorf("valor da recarga deve ser positivo: %.2f", valor)
	}

	path := filepath.Join("app", "internal", "dataJson", "veiculos.json")

	alternatives := []string{
		"internal/dataJson/veiculos.json",
		"veiculos.json",
		"./veiculos.json",
		"/app/internal/dataJson/veiculos.json",
	}

	// Ler arquivo existente
	file, err := os.Open(path)

	// Se falhar, tentar caminhos alternativos
	if err != nil {
		var lastErr error = err
		for _, altPath := range alternatives {
			file, err = os.Open(altPath)
			if err == nil {
				path = altPath
				break
			}
			lastErr = err
		}

		// Se todos os caminhos falharem, criar um novo arquivo
		if err != nil {
			// Criar um novo arquivo com este veículo
			dadosVeiculos := DadosVeiculos{
				Veiculos: []Veiculo{
					{
						Placa: placa,
						Recargas: []Recarga{
							{
								Data:    time.Now().Format("2006-01-02 15:04:05"),
								PontoID: pontoID,
								Valor:   valor,
							},
						},
					},
				},
			}

			// Tentar salvar em todos os caminhos possíveis
			for _, savePath := range append([]string{path}, alternatives...) {
				err = salvarDadosVeiculos(savePath, dadosVeiculos)
				if err == nil {
					return nil // Salvou com sucesso
				}
			}
			return fmt.Errorf("não foi possível criar arquivo de veículos: %v", lastErr)
		}
	}
	defer file.Close()

	// Decodificar dados existentes
	var dadosVeiculos DadosVeiculos
	if err := json.NewDecoder(file).Decode(&dadosVeiculos); err != nil {
		// Se não conseguir decodificar, criar uma nova estrutura
		dadosVeiculos = DadosVeiculos{
			Veiculos: []Veiculo{},
		}
	}

	// Data atual
	dataAtual := time.Now().Format("2006-01-02 15:04:05")

	// Encontrar veículo e adicionar recarga
	encontrado := false
	for i, v := range dadosVeiculos.Veiculos {
		if v.Placa == placa {
			novaRecarga := Recarga{
				Data:    dataAtual,
				PontoID: pontoID,
				Valor:   valor,
			}
			dadosVeiculos.Veiculos[i].Recargas = append(dadosVeiculos.Veiculos[i].Recargas, novaRecarga)
			encontrado = true
			break
		}
	}

	if !encontrado {
		// Adicionar novo veículo com esta recarga
		dadosVeiculos.Veiculos = append(dadosVeiculos.Veiculos, Veiculo{
			Placa: placa,
			Recargas: []Recarga{
				{
					Data:    dataAtual,
					PontoID: pontoID,
					Valor:   valor,
				},
			},
		})
	}

	return salvarDadosVeiculos(path, dadosVeiculos)
}

func salvarDadosVeiculos(path string, dados DadosVeiculos) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(dados)
}

func ObterUltimoReserva(placa string) (int, error) {
	path := filepath.Join("app", "internal", "dataJson", "veiculos.json")

	// Ler arquivo existente
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Decodificar dados existentes
	var dadosVeiculos DadosVeiculos
	if err := json.NewDecoder(file).Decode(&dadosVeiculos); err != nil {
		return 0, err
	}

	// Encontrar veículo e obter a última reserva
	for _, v := range dadosVeiculos.Veiculos {
		if v.Placa == placa && len(v.Recargas) > 0 {
			// Retornar o ponto da recarga mais recente
			// Assumindo que a última posição do array é a mais recente
			return v.Recargas[len(v.Recargas)-1].PontoID, nil
		}
	}

	return 0, fmt.Errorf("veículo com placa %s não encontrado ou sem reservas", placa)
}

func PlacaJaExiste(placa string) bool {
	path := filepath.Join("app", "internal", "dataJson", "veiculos.json")

	// Tentar abrir o arquivo
	file, err := os.Open(path)
	if err != nil {
		// Se o arquivo não existe, a placa não existe
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	defer file.Close()

	// Decodificar dados existentes
	var dadosVeiculos DadosVeiculos
	if err := json.NewDecoder(file).Decode(&dadosVeiculos); err != nil {
		return false
	}

	// Verificar se a placa já existe
	for _, v := range dadosVeiculos.Veiculos {
		if v.Placa == placa {
			return true // Placa já existe
		}
	}

	return false // Placa não existe
}

func ObterHistoricoRecargas(placa string) ([]Recarga, error) {
	path := filepath.Join("app", "internal", "dataJson", "veiculos.json")
	alternatives := []string{
		"internal/dataJson/veiculos.json",
		"veiculos.json",
		"./veiculos.json",
		"/app/internal/dataJson/veiculos.json",
	}

	// Tentar abrir o arquivo
	file, err := os.Open(path)

	// Se falhar, tentar caminhos alternativos
	if err != nil {
		for _, altPath := range alternatives {
			file, err = os.Open(altPath)
			if err == nil {
				break
			}
		}

		if err != nil {
			// Se todos os caminhos falharem, não há histórico
			return []Recarga{}, nil
		}
	}
	defer file.Close()

	// Decodificar dados existentes
	var dadosVeiculos DadosVeiculos
	if err := json.NewDecoder(file).Decode(&dadosVeiculos); err != nil {
		return nil, err
	}

	// Procurar o veículo pela placa
	for _, v := range dadosVeiculos.Veiculos {
		if v.Placa == placa {
			return v.Recargas, nil
		}
	}

	// Veículo não encontrado
	return []Recarga{}, nil
}
