package main

import (
	"fmt"
	"log"
	"net"
	"recarga-inteligente/internal/handler"
)

func main() {
	fmt.Println("Inicializando servidor")

	//Inicializa o servidor na porta 5000
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
	defer listener.Close()

	fmt.Println("Servidor escutando na porta 5000...")

	//Aguarda conexoes
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Erro ao aceitar conexao: %v", err)
			continue
		}

		//Trata a conexao em uma goroutine separada
		go handler.HandleConnection(conn)
	}
}
