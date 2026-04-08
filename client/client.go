package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const BROKER_ADDR = "192.168.0.103:9091"

var (
	conn               net.Conn
	leitor             *bufio.Reader
	nick               string
	sensorAtual        string
	aguardandoResposta = make(chan string)
)

// ── estilos ───────────────────────────────────────────────────────────────────
var (
	corVerde    = lipgloss.Color("#00FF87")
	corAzul     = lipgloss.Color("#00B4D8")
	corAmarelo  = lipgloss.Color("#FFD60A")
	corVermelho = lipgloss.Color("#FF4D4D")
	corCinza    = lipgloss.Color("#6C757D")
	corBranco   = lipgloss.Color("#FFFFFF")

	estiloCaixa = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(corAzul).
			Padding(0, 2)

	estiloTitulo = lipgloss.NewStyle().
			Bold(true).
			Foreground(corAzul).
			Align(lipgloss.Center).
			Width(40)

	estiloSubtitulo = lipgloss.NewStyle().
			Foreground(corCinza).
			Align(lipgloss.Center).
			Width(40)

	estiloNumero = lipgloss.NewStyle().
			Bold(true).
			Foreground(corAzul)

	estiloOpcao = lipgloss.NewStyle().
			Foreground(corBranco)

	estiloSensorTag = lipgloss.NewStyle().
			Bold(true).
			Foreground(corVerde).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(corVerde).
			PaddingLeft(1)

	estiloSensorDado = lipgloss.NewStyle().
				Foreground(corVerde)

	estiloBroker = lipgloss.NewStyle().
			Foreground(corAmarelo)

	estiloErro = lipgloss.NewStyle().
			Foreground(corVermelho).
			Bold(true)

	estiloSucesso = lipgloss.NewStyle().
			Foreground(corVerde).
			Bold(true)

	estiloPrompt = lipgloss.NewStyle().
			Foreground(corAzul).
			Bold(true)

	estiloSecao = lipgloss.NewStyle().
			Foreground(corAzul).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(corAzul).
			Width(40).
			MarginBottom(1)
)

// ─────────────────────────────────────────────────────────────────────────────

func getBrokerAddr() string {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		return BROKER_ADDR
	}
	return addr
}

func limparTela() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func pausar(scanner *bufio.Scanner) {
	fmt.Print(estiloPrompt.Render("\nPressione ENTER para voltar ao menu..."))
	scanner.Scan()
}

func main() {
	var err error
	conn, err = net.Dial("tcp", getBrokerAddr())
	if err != nil {
		fmt.Println(estiloErro.Render("✗ Erro ao conectar no broker: " + err.Error()))
		os.Exit(1)
	}
	defer conn.Close()

	leitor = bufio.NewReader(conn)
	scanner := bufio.NewScanner(os.Stdin)

	limparTela()
	fmt.Println(estiloCaixa.Render(
		estiloTitulo.Render("🌐  BROKER IoT") + "\n" +
			estiloSubtitulo.Render("Sistema de Integração para Dispositivos IoT"),
	))

	fmt.Print(estiloPrompt.Render("  Digite seu nick: "))
	scanner.Scan()
	nick = strings.TrimSpace(scanner.Text())

	enviar("REGISTER-CLIENT " + nick)
	resp, _ := leitor.ReadString('\n')
	if !strings.Contains(resp, "CONECTADO") {
		log.Fatalln("Erro ao registrar:", resp)
	}

	fmt.Println(estiloSucesso.Render("  ✓ Conectado como: " + nick))

	go escutarBroker()

	for {
		limparTela()
		mostrarMenu()
		fmt.Print(estiloPrompt.Render("\n  Escolha uma opção → "))

		if !scanner.Scan() {
			break
		}
		opcao := strings.TrimSpace(scanner.Text())
		limparTela()

		switch opcao {
		case "1":
			fmt.Println(estiloSecao.Render("  LISTA DE SENSORES"))
			enviar("LIST-SENSORES")
			<-aguardandoResposta
			pausar(scanner)

		case "2":
			fmt.Println(estiloSecao.Render("  SEGUIR SENSOR"))
			fmt.Print(estiloPrompt.Render("  Nick do sensor: "))
			scanner.Scan()
			ns := strings.TrimSpace(scanner.Text())
			seguirSensor(ns)
			pausar(scanner)

		case "3":
			fmt.Println(estiloSecao.Render("  PARAR SENSOR"))
			pararSensor()
			pausar(scanner)

		case "4":
			fmt.Println(estiloSecao.Render("  LISTA DE ATUADORES"))
			enviar("LIST-ATUADORES")
			<-aguardandoResposta
			pausar(scanner)

		case "5":
			fmt.Println(estiloSecao.Render("  COMANDAR ATUADOR"))
			comandarAtuador(scanner)
			pausar(scanner)

		case "6":
			fmt.Println(estiloSecao.Render("  LISTA DE CLIENTES"))
			enviar("LIST-CLIENTES")
			<-aguardandoResposta
			pausar(scanner)

		case "7":
			enviar("QUIT")
			limparTela()
			fmt.Println(estiloTitulo.Render("👋  Até logo, " + nick + "!"))
			os.Exit(0)

		default:
			fmt.Println(estiloErro.Render("  ✗ Opção inválida! Tente novamente."))
			pausar(scanner)
		}
	}
}

func mostrarMenu() {
	statusSensor := ""
	if sensorAtual != "" {
		statusSensor = "\n" + estiloSensorTag.Render("📡 Ouvindo: "+sensorAtual)
	}

	opcoes := []struct{ num, txt string }{
		{"1", "Listar sensores"},
		{"2", "Seguir sensor"},
		{"3", "Parar de seguir sensor"},
		{"4", "Listar atuadores"},
		{"5", "Comandar atuador"},
		{"6", "Listar clientes"},
		{"7", "Sair"},
	}

	linhas := ""
	for _, o := range opcoes {
		linhas += "  " + estiloNumero.Render("["+o.num+"]") + " " + estiloOpcao.Render(o.txt) + "\n"
	}

	conteudo := estiloTitulo.Render("🌐  BROKER IoT - CLIENTE") + "\n" +
		estiloSubtitulo.Render("conectado como: "+nick) +
		statusSensor + "\n\n" +
		strings.TrimRight(linhas, "\n")

	fmt.Println(estiloCaixa.Render(conteudo))
}

func seguirSensor(nickSensor string) {
	if sensorAtual != "" {
		fmt.Println(estiloErro.Render("  ✗ Já seguindo " + sensorAtual + ". Pare primeiro (opção 3)."))
		return
	}
	enviar("SEGUIR-SENSOR " + nickSensor)
	resp := <-aguardandoResposta
	if strings.Contains(strings.ToLower(resp), "ok") {
		sensorAtual = nickSensor
		fmt.Println(estiloSucesso.Render("  ✓ Seguindo: " + nickSensor))
	} else {
		fmt.Println(estiloErro.Render("  ✗ " + resp))
	}
}

func pararSensor() {
	if sensorAtual == "" {
		fmt.Println(estiloErro.Render("  ✗ Nenhum sensor ativo."))
		return
	}
	enviar("PARAR-SENSOR " + sensorAtual)
	<-aguardandoResposta
	fmt.Println(lipgloss.NewStyle().Foreground(corAmarelo).Bold(true).Render("  ✓ Parou de seguir: " + sensorAtual))
	sensorAtual = ""
}

func comandarAtuador(scanner *bufio.Scanner) {
	fmt.Print(estiloPrompt.Render("  Nick do atuador: "))
	scanner.Scan()
	id := strings.TrimSpace(scanner.Text())

	fmt.Print(estiloPrompt.Render("  Comando (ON/OFF): "))
	scanner.Scan()
	cmd := strings.TrimSpace(scanner.Text())

	enviar(fmt.Sprintf("COMMAND %s %s", id, cmd))
	<-aguardandoResposta
}

func enviar(msg string) {
	conn.Write([]byte(msg + "\n"))
}

func escutarBroker() {
	for {
		msg, err := leitor.ReadString('\n')
		if err != nil {
			fmt.Println(estiloErro.Render("\n  ✗ Conexão perdida."))
			os.Exit(0)
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}

		if sensorAtual != "" && strings.HasPrefix(msg, sensorAtual+":") {
			fmt.Printf("\r%s", estiloSensorDado.Render("  📡 "+msg+strings.Repeat(" ", 20)))
		} else {
			fmt.Printf("\n%s\n", estiloBroker.Render("  ℹ  "+msg))
			select {
			case aguardandoResposta <- msg:
			default:
			}
		}
	}
}