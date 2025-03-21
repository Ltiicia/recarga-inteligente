package handler

import (
	"fmt"
	"net"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	//Recebe e le os dados da conexao
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Erro ao ler dados: %v", err)
		return
	}

	//Exibe a mensagem recebida
	mensagem := string(buffer[:n])
	fmt.Printf("Mensagem recebida: %s", mensagem)

	//Responde
	resposta := "Mensagem recebida com sucesso!"
	_, err = conn.Write([]byte(resposta))
	if err != nil {
		fmt.Printf("Erro ao enviar resposta: %v", err)
	}
}
