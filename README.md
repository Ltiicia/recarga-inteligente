<h2 align="center">Recarga de carros elétricos inteligente</h2>
<h4 align="center">Projeto da disciplina TEC502 - Concorrência e Conectividade.</h4>

<p align="center">Este projeto foi desenvolvido para facilitar a comunicação entre veículos elétricos e pontos de recarga. Utilizando uma arquitetura cliente-servidor baseada no protocolo TCP/IP, o sistema permite que veículos solicitem recargas, informem sua localização atual gerada de forma randômica e recebam recomendações para pontos de recarga próximos.</p>
<p align="center">O objetivo é otimizar o processo de recarga, garantindo eficiência e gerenciamento adequado da concorrência.</p>

## Sumário

- [Introdução](#introdução)
- [Funcionalidades](#funcionalidades)
- [Arquitetura do Sistema](#arquitetura-do-sistema)
- [Protocolo de Comunicação](#protocolo-de-comunicação)
- [Gerenciamento de Concorrência](#gerenciamento-de-concorrência)
- [Como Executar](#como-executar)
- ...
- [Referências](#referências)

## Introdução

O presente sistema foi desenvolvido para implementar comunicação entre cliente-servidor simulando o contexto de recarga de veículos elétricos. O projeto viabiliza a solicitação e gestão de recargas por parte dos veículos, utilizando o protocolo TCP/IP e desenvolvimento em Go, com suporte para múltiplas conexões simultâneas.  

A aplicação está contida em containers Docker, que isolam e orquestram a execução dos serviços. Onde:
- O servidor gerencia os pontos de recarga disponíveis, recebe solicitações dos veículos, calcula distâncias, gerencia as filas e administra as reservas. Ele é responsável por validar as transações de recarga, verificando a disponibilidade dos pontos, e tratando o armazenamento das informações. 
- O veículo, por sua vez, permite ao usuário solicitar recargas, informa sua localização atual para consultar pontos de recarga disponíveis e escolher onde realizar a operação. 
- Já o ponto de recarga, é responsável por conectar-se ao servidor quando estiver disponível para realização de recargas. Informando a sua disponibilidade ou fila de espera e gerenciando localmente sua fila de reservas. Ao receber uma reserva, o ponto de recarga processa o atendimento ao veículo, atualiza sua fila e libera o ponto após a conclusão do carregamento.    

Porcionando então, uma solução que permite aos veículos encontrar, reservar e utilizar pontos de recarga de forma otimizada, considerando fatores como proximidade e disponibilidade.  

## Arquitetura do Sistema

A solução foi desenvolvida utilizando a arquitetura de comunicação cliente-servidor, onde a comunicação entre as partes ocorre por meio do protocolo Transmission Control Protocol (TCP). Seu uso garante a integridade e ordem dos pacotes proporcionando uma comunicação confiável entre os módulos do sistema: servidor, veículos e pontos de recarga. 

Toda a troca de dados ocorre via conexões TCP/IP, com mensagens estruturadas em formato JSON. O sistema foi projetado para funcionar em ambiente de containers Docker interconectados por uma rede interna definida no docker-compose, garantindo isolamento, escalabilidade e simulação de concorrência. Onde:

- **Servidor**: Gerencia as solicitações, consulta os pontos, calcula distâncias e gerenciar as filas.
- **Veículo**: Responsável por solicitar recargas, informar sua localização e escolher o ponto de recarga.
- **Ponto de Recarga**: Mantém sua fila local e responde às requisições de disponibilidade do servidor.

### Servidor
O servidor atua como o núcleo central do sistema, responsável por intermediar a comunicação entre veículos e pontos de recarga, escutando conexões TCP em uma porta definida. As principais responsabilidades do servidor incluem:
- Gerenciar conexões TCP de veículos e pontos de recarga.
- Gerenciar solicitações de recarga dos veículos, calcular a melhor opção com base em distância e fila de espera e apresentar as 3 melhores alternativas.
- Gerenciar as reservas, garantindo que cada veículo seja corretamente adicionado à fila de um ponto selecionado.  
O servidor foi desenvolvido em Go, utilizando recursos como goroutines para o tratamento concorrente de conexões e channels para comunicação entre rotinas. Isso garante maior performance e segurança no acesso aos dados compartilhados.

### Ponto de Recarga
Cada ponto de recarga é implementado como um cliente TCP. Inicialmente, o sistema possui 8 pontos de recarga previamente cadastrados que podem se conectar simultaneamente. Cada ponto ao se conectar se identifica, permitindo que o servidor o associe aos dados cadastrados em um arquivo json, contendo sua localização geográfica, sendo identificado por um ID único e mantém comunicação contínua para:
- Enviar sua disponibilidade / fila atual de veículos aguardando recarga.
- Gerenciar localmente sua fila de reservas
- Processar o atendimento ao veículo  
Cada ponto gerencia localmente sua própria fila e responde dinamicamente a solicitações do servidor. Caso um ponto seja desconectado, seu ID é liberado automaticamente pelo servidor, permitindo a reutilização por novas conexões.

### Veículo
O veículo também é implementado como cliente TCP onde o usuário interage por meio de um menu via terminal que permite:
- Enviar sua localização atual ao solicitar recarga.
- Receber sugestões de pontos disponíveis, com fila de espera e distância.
- Escolher um ponto de recarga para reservar e efetuar recarga  
O sistema é capaz de manter sessões interativas com o servidor, permitindo que o usuário envie solicitações de recarga e consulte seu histórico de recargas pendentes para efetuar o pagamento posteriormente.  

A comunicação entre as partes ocorre via **sockets TCP/IP** conforme ilustração da arquitetura à seguir:

<div align="center">  
  <img align="center" width=100% src= public/sistema-recarga.png alt="Comunicação sistema">
  <p><em>Arquitetura do Sistema</em></p>
</div>

### Comunicação

- Veículo solicita recarga ao servidor.
- Servidor solicita localização atual ao veículo.
- Veículo envia sua localização (latitude e longitude) atual.
- Servidor solicita disponibilidade/fila aos pontos de recarga conectados.
- Pontos enviam disponibilidade/fila atual.
- Servidor calcula as distâncias do veículo até os pontos, verifica fila e define as melhores opções.
- Servidor envia melhores opções ao veículo.
- Veículo seleciona um ponto e solicita reserva.
- Servidor confirma a reserva e adiciona o veículo à fila do ponto.
- Veículo se desloca e realiza a recarga.
- Ponto remove o veículo da sua fila ao final da recarga.
- O valor da recarga é vinculado ao veículo.

### Funcionalidades Principais

- **Solicitação de Recarga**: O veículo pode solicitar uma recarga ao servidor.
- **Envio de Localização**: O servidor solicita e recebe a localização atual do veículo, gerada aleatoriamente.
- **Consulta de Disponibilidade**: O servidor consulta os pontos de recarga conectados sobre sua disponibilidade ou fila de espera.
- **Cálculo de Distância**: O servidor calcula a distância entre o veículo e os pontos de recarga disponíveis.
- **Reserva de Ponto de Recarga**: O veículo recebe as opções e seleciona o ponto desejado.
- **Gerenciamento de Fila**: O servidor efetua a reserva adicionando o veículo à fila do ponto de recarga escolhido.
- **Finalização e Liberação**: O veículo é removido da fila ao final da recarga e recebe o valor para pagamento.

## Protocolo de Comunicação
A comunicação entre os clientes e o servidor é baseada em mensagens JSON transmitidas via sockets TCP. O formato JSON foi escolhido por ser leve, legível e amplamente adotado em sistemas distribuídos. Cada mensagem permite a troca de informações, dados, além de encapsular ações como identificação, requisição de recarga, resposta de disponibilidade, reservas, entre outros.

### Dados e Estado
Os dados do sistema como área de cobertura e localização dos pontos de recarga cadastrados, são carregados a partir de arquivos JSON ao iniciar o servidor e permanecem em memória, funcionando como um cache de alta performance. Isso reduz a latência e permite respostas rápidas às requisições dos veículos.  

### Paradigma de Comunicação
O sistema segue o paradigma stateless para comunicação com os veículos. Cada requisição é independente, e os veículos devem se identificar a cada nova interação. Isso facilita a escalabilidade e permite a execução de múltiplas instâncias do servidor, se necessário. Apesar de ser stateless para veículos, o sistema possui uma gestão de estado parcial para os pontos de recarga, que mantêm sessões ativas enquanto conectados ao servidor, permitindo a troca contínua de mensagens e atualização de suas filas de recarga.

## Conexões Simultâneas
O servidor foi projetado para suportar múltiplas conexões simultâneas utilizando goroutines, nativas da linguagem Go. A cada nova conexão, uma nova goroutine é iniciada, permitindo que o servidor processe requisições de forma paralela e responsiva, sem bloquear outras conexões, maximizando a escalabilidade do sistema e garantindo que a resposta a uma solicitação de recarga, por exemplo, não afete outras conexões ativas.

## Gerenciamento de Concorrência
Para garantir a integridade dos dados durante operações concorrentes como por exemplo a inserção em filas de espera para reserva de recargas, foi implementado o uso de mutexes. O controle de exclusão mútua assegura que múltiplas goroutines não modifiquem simultaneamente estruturas de dados compartilhadas, como a fila de espera de um ponto de recarga.  

Funcionamento:  
- Lock: Antes da operação crítica, a goroutine realiza um mutex.Lock().  
- Seção Crítica: Os dados são validados e atualizados de forma segura.
- Unlock: Após a operação, o mutex é liberado com mutex.Unlock(), permitindo que outras goroutines prossigam.  

Essa abordagem evita problemas como múltiplos veículos tentando ocupar a mesma posição na fila de reservas de um determinado ponto de recarga.

### Garantia de Reserva e Integridade
Ao solicitar uma recarga, o veículo envia sua localização atual ao servidor. O servidor, então:

- Solicita a disponibilidade / fila dos pontos de recarga conectados.
- Calcula as distâncias e os scores estimados com base nas filas.
- Retorna ao veículo as três melhores opções de pontos.

Após a escolha, o veículo é adicionado à fila do ponto selecionado. Para garantir a integridade da operação, cada etapa é realizada com controle de concorrência utilizando mutexes e channels, impedindo que dois veículos reservem a mesma posição simultaneamente.

### Execução com Docker
A simulação do sistema é feita utilizando docker-compose, com containers para o Servidor, os Pontos de recarga e os Veículos. O Docker Compose permite aos módulos compartilhar uma rede interna privada, proporcionando a troca de mensagens TCP entre os containers.

## Tecnologias Utilizadas
- Linguagem: Go (Golang)
- Comunicação: TCP/IP com net.Conn
- Contêineres: Docker, Docker Compose
- Mock de dados: JSON

## Como Executar
...

## Desenvolvedoras
<table>
  <tr>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/142849685?v=4" width="100px;" alt=""/><br /><sub><b> Brenda Araújo </b></sub></a><br />👨‍💻</a></td>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/89545660?v=4" width="100px;" alt=""/><br /><sub><b> Naylane Ribeiro </b></sub></a><br />👨‍💻</a></td>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/124190885?v=4" width="100px;" alt=""/><br /><sub><b> Letícia Gonçalves </b></sub></a><br />👨‍💻</a></td>    
  </tr>
</table>

## Referências
