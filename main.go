package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"runtime"
)

const (
	BUFFER_SIZE = 2048
)

type Config struct {
	Host    string
	Servers []string
}

var verbose = flag.Bool("v", false, "verbose output")
var help = flag.Bool("h", false, "help message")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] godist.conf\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()

	if *help {
		flag.Usage()
	}

	var config *Config
	if flag.NArg() == 1 {
		config = readConfigFile(flag.Arg(0))
	} else if flag.NArg() == 0 {
		config = readConfigFile("godist.conf")
	} else {
		flag.Usage()
	}

	nCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nCPU)

	proxyServer(config)
}

func proxyServer(config *Config) {
	listener, err := net.Listen("tcp", config.Host)
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("Starting godist on " + config.Host)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}

		assignedServer := serverFromString(config, conn.RemoteAddr().String())
		go proxyConn(conn, assignedServer)
	}
}

func proxyConn(clientConn net.Conn, serverURL string) {

	tcpAddr, err := net.ResolveTCPAddr("tcp", serverURL)
	if err != nil {
		log.Panicln(err.Error())
	}

	serverConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Panicln("Could not connect to server " + serverURL)
		log.Panicln(err.Error())
	}

	if *verbose {
		log.Println("New proxy session with client " + clientConn.RemoteAddr().String() + " and server " + serverConn.RemoteAddr().String())
	}

	finish := make(chan bool, 2)

	go passBytes(clientConn, serverConn, finish)
	go passBytes(serverConn, clientConn, finish)

	<-finish
	clientConn.Close()
	serverConn.Close()
	<-finish

	if *verbose {
		log.Println("Proxy session closed with client " + clientConn.RemoteAddr().String() + " and server " + serverConn.RemoteAddr().String())
	}
}

func passBytes(from, to net.Conn, finish chan bool) {
	defer func() { finish <- true }()

	for {
		buf := new(bytes.Buffer)
		for {
			data := make([]byte, BUFFER_SIZE)
			n, err := from.Read(data)
			if err != nil {
				if *verbose {
					log.Println(from.RemoteAddr().String() + " closed connection with " + to.RemoteAddr().String() + " — " + err.Error())
				}
				return
			}

			buf.Write(data[:n])
			if data[len(data)-1] == 0 {
				if _, err := to.Write(buf.Bytes()); err != nil {
					if *verbose {
						log.Println(to.RemoteAddr().String() + " closed connection with " + from.RemoteAddr().String() + " — " + err.Error())
					}
					return
				}

				if *verbose {
					log.Printf("%v bytes sent from "+from.RemoteAddr().String()+" to "+to.RemoteAddr().String()+"\n", len(buf.Bytes()))
				}

				break
			}
		}
	}

}

func serverFromString(config *Config, str string) string {
	index := int(math.Floor(math.Mod(float64(hash(str)), float64(len(config.Servers)))))

	return config.Servers[index]
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))

	return int(h.Sum32())
}

func readConfigFile(file string) *Config {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fatal(err.Error())
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fatal(err.Error())
	}

	return &config
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
