package main

import (
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

type config struct {
	host    string
	servers []string
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

	var conf *config
	if flag.NArg() == 1 {
		conf = readConfigFile(flag.Arg(0))
	} else if flag.NArg() == 0 {
		conf = readConfigFile("godist.conf")
	} else {
		flag.Usage()
	}

	nCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nCPU)

	proxyServer(conf)
}

func proxyServer(conf *config) {
	listener, err := net.Listen("tcp", conf.host)
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("Starting godist on " + conf.host)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}

		assignedServer := serverFromString(conf, conn.RemoteAddr().String())
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
		data := make([]byte, BUFFER_SIZE)

		n, err := from.Read(data)
		if err != nil {
			if *verbose {
				log.Println(from.RemoteAddr().String() + " closed connection with " + to.RemoteAddr().String() + " — " + err.Error())
			}
			return
		}

		if _, err := to.Write(data[:n]); err != nil {
			if *verbose {
				log.Println(to.RemoteAddr().String() + " closed connection with " + from.RemoteAddr().String() + " — " + err.Error())
			}
			return
		}

		if *verbose {
			log.Printf("%v bytes sent from "+from.RemoteAddr().String()+" to "+to.RemoteAddr().String()+"\n", n)
		}
	}

}

func serverFromString(conf *config, str string) string {
	index := int(math.Floor(math.Mod(float64(hash(str)), float64(len(conf.servers)))))

	return conf.servers[index]
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))

	return int(h.Sum32())
}

func readConfigFile(file string) *config {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fatal(err.Error())
	}

	var conf config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		fatal(err.Error())
	}

	return &conf
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
