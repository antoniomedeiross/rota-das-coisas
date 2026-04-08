# Broker IoT

Sistema distribuído em Go para comunicação entre **sensores**, **atuadores** e **clientes**, com um **broker** central responsável por registrar dispositivos, receber dados, encaminhar comandos e listar conexões.

## Visão geral

O projeto é dividido em quatro partes principais:

- **Broker**: núcleo do sistema, recebe conexões UDP e TCP.
- **Sensores**: enviam dados periódicos via UDP.
- **Atuadores**: recebem comandos via TCP.
- **Clientes**: interface interativa para consultar sensores/atuadores/clientes e comandar atuadores.

## Estrutura de diretórios

- [broker/](broker)
  - [broker/main.go](broker/main.go)
  - [broker/config/config.go](broker/config/config.go)
  - [broker/handler/tcp_handler.go](broker/handler/tcp_handler.go)
  - [broker/handler/udp_handler.go](broker/handler/udp_handler.go)
  - [broker/model/models.go](broker/model/models.go)
  - [broker/netServer/tcp.go](broker/netServer/tcp.go)
  - [broker/netServer/udp.go](broker/netServer/udp.go)
  - [broker/repository/repository.go](broker/repository/repository.go)
  - [broker/repository/heartBeat.go](broker/repository/heartBeat.go)
- [client/](client)
  - [client/client.go](client/client.go)
  - [client/client-cla.go](client/client-cla.go)
  - [client/client-gem.go](client/client-gem.go)
  - [client/go.mod](client/go.mod)
- [sensores/](sensores)
  - [sensores/main.go](sensores/main.go)
- [atuadores/](atuadores)
  - [atuadores/atuador.go](atuadores/atuador.go)
- [test/](test)
  - [test/test-concorrencia.go](test/test-concorrencia.go)

## Pacotes do broker

### [broker/config/config.go](broker/config/config.go)
Centraliza as portas e IPs padrão do broker.

- UDP: `9090`
- TCP: `9091`

### [broker/model/models.go](broker/model/models.go)
Define as estruturas de dados do sistema:

- `Sensor`
- `Atuador`
- `Cliente`

### [broker/repository/repository.go](broker/repository/repository.go)
Responsável pelo armazenamento em memória e pelas operações principais:

- registro de sensores, atuadores e clientes
- listagem de dispositivos conectados
- comando de atuadores
- inscrição e cancelamento de sensores
- repasse de dados de sensores para clientes inscritos

### [broker/repository/heartBeat.go](broker/repository/heartBeat.go)
Monitora sensores por tempo de inatividade e remove sensores antigos da lista.

### [broker/handler/tcp_handler.go](broker/handler/tcp_handler.go)
Processa comandos TCP, como:

- `REGISTER-CLIENT`
- `REGISTER-ATUADOR`
- `COMMAND`
- `LIST-SENSORES`
- `LIST-ATUADORES`
- `LIST-CLIENTES`
- `SEGUIR-SENSOR`
- `PARAR-SENSOR`
- `QUIT`

### [broker/handler/udp_handler.go](broker/handler/udp_handler.go)
Processa mensagens UDP:

- `REGISTER-SENSOR`
- `DATA`
- `HELP`

### [broker/netServer/tcp.go](broker/netServer/tcp.go)
Sobe o servidor TCP na porta `9091`.

### [broker/netServer/udp.go](broker/netServer/udp.go)
Sobe o servidor UDP na porta `9090`.

### [broker/main.go](broker/main.go)
Inicializa os servidores TCP e UDP e o monitoramento de heartbeat.

## Como executar

## Com Docker

### Broker
```sh
docker build -t broker ./broker
docker run --rm -p 9090:9090/udp -p 9091:9091/tcp broker
```

### Cliente
```sh
cd client
docker build -t client -f Dockerfile.client .
docker run --rm -it -e SERVER_ADDR=<IP_DO_BROKER>:9091 client
```

### Sensores
```sh
cd sensores
docker build -t sensores -f Dockerfile.sensores .
docker run --rm -it -e SERVER_ADDR=<IP_DO_BROKER> sensores
```

### Atuadores
```sh
cd atuadores
docker build -t atuadores -f Dockerfile.atuadores .
docker run --rm -it -e SERVER_ADDR=<IP_DO_BROKER> atuadores
```

## Com Docker Compose

Cada pasta possui um arquivo `docker-compose.yml`:

- [broker/docker-compose.yml](broker/docker-compose.yml)
- [client/docker-compose.yml](client/docker-compose.yml)
- [sensores/docker-compose.yml](sensores/docker-compose.yml)
- [atuadores/docker-compose.yml](atuadores/docker-compose.yml)

Exemplo:

```sh
cd broker
docker compose up --build
```


## Como usar

## Cliente
Ao iniciar, o cliente solicita o nick e exibe um menu interativo:

1. Listar sensores
2. Seguir sensor
3. Parar de seguir sensor
4. Listar atuadores
5. Comandar atuador
6. Listar clientes
7. Sair

### Exemplos de uso
- Seguir um sensor:
  - escolher a opção `2`
  - informar o nick do sensor
- Comandar um atuador:
  - escolher a opção `5`
  - informar o nick do atuador
  - informar `ON` ou `OFF`

## Sensores
O sensor:

- registra-se no broker via UDP
- envia dados periódicos com `DATA <nick> <valor>`
- mantém envio contínuo em intervalos curtos

## Atuadores
O atuador:

- registra-se no broker via TCP
- aguarda comandos `ON` e `OFF`
- responde com o estado atual

## Teste de concorrência

O teste está em [test/test-concorrencia.go](test/test-concorrencia.go).

Executar:
```sh
cd test
go run test-concorrencia.go
```

Esse teste:

- abre dois clientes simultâneos
- envia comandos repetidos para o mesmo atuador
- mede o tempo de resposta

## Dependências

- Go `1.22`
- `github.com/charmbracelet/lipgloss`

## Observações

- O sistema usa **UDP** para sensores e **TCP** para clientes e atuadores.
- Sensores clientes e atuadores precisam apontar para o **IP do broker** via variável `SERVER_ADDR`.
- O cliente usa a porta TCP `9091`.


## Licença

Projeto distribuído sob a licença MIT. Veja [LICENSE](LICENSE).