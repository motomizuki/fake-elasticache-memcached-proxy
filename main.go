package main

import (
	"bufio"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	host               string
	port               string
	proxyMemcachedHost string
	proxyMemcachedPort string
	clusterNodes       string
	trace              bool
}

type Handler struct {
	conn  net.Conn
	addr  string
	nodes string
	trace bool
}

type Config struct {
	Host               string `default:"0.0.0.0"`
	Port               string `default:"11211"`
	ProxyMemcachedHost string `split_words:"true" required:"true"`
	ProxyMemcachedPort string `split_words:"true" required:"true"`
	ClusterNodes       string `split_words:"true" default:"localhost|127.0.0.1|11211"`
	Trace              bool   `default:"false"`
}

func New(config *Config) *Server {
	return &Server{
		host:               config.Host,
		port:               config.Port,
		proxyMemcachedHost: config.ProxyMemcachedHost,
		proxyMemcachedPort: config.ProxyMemcachedPort,
		clusterNodes:       config.ClusterNodes,
		trace:              config.Trace,
	}
}

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.host, server.port))
	proxyAddr := fmt.Sprintf("%s:%s", server.proxyMemcachedHost, server.proxyMemcachedPort)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		handler := &Handler{
			conn:  conn,
			addr:  proxyAddr,
			nodes: server.clusterNodes,
			trace: server.trace,
		}

		go handler.handleRequest()
	}
}
func (handler *Handler) handleConfigCommand(message string) []byte {
	nodes := handler.nodes + "\n\r\n"
	size := strconv.Itoa(len(nodes))
	// response format: CONFIG cluster 0 ${byte size} \r\n${version}\n${host}|${ip}|${port}\n\r\nEND\r\n
	response := []byte("CONFIG cluster 0 " + size + "\r\n5\n" + nodes + "END\r\n")

	if handler.trace {
		log.Printf("trace { request=%#v, response=%#v }\n", message, string(response))
	}

	return response
}

func (handler *Handler) handleMemachedCommand(message string, reader *bufio.Reader) []byte {
	// proxy request to Memcached
	mc, err := net.Dial("tcp", handler.addr)

	if err != nil {
		log.Println(err)
	}

	command := []byte(message)

	if strings.HasPrefix(message, "set") {
		// set is a multiple line command so read additional line
		value, _ := reader.ReadBytes(byte(0x0a))
		command = append(command, value...)
	}

	mc.Write(command)

	// proxy response from Memcached
	mReader := bufio.NewReader(mc)
	var response []byte

	for {
		ret, _ := mReader.ReadString(byte(0x0a))

		if strings.HasPrefix(ret, "VALUE") {
			// The type of VALUE is multiple line response, so read additional bytes.
			s := strings.Split(ret, " ")
			vs, _ := strconv.Atoi(strings.Trim(s[3], "\r\n"))
			val := make([]byte, vs+2)
			mReader.Read(val)
			response = append(response, []byte(ret)...)
			response = append(response, val...)
			// The type of VALUE is returned multiple times until END arrives, so continue the loop
			continue
		} else {
			response = append(response, []byte(ret)...)
			break
		}
	}

	mc.Close()

	if handler.trace {
		log.Printf("trace { request=%#v, response=%#v }\n", string(command), string(response))
	}

	return response
}

func (handler *Handler) handleRequest() {
	reader := bufio.NewReader(handler.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			handler.conn.Close()
			return
		}
		var response []byte
		if message == "config get cluster\r\n" {
			response = handler.handleConfigCommand(message)
		} else {
			response = handler.handleMemachedCommand(message, reader)
		}
		handler.conn.Write(response)
	}
}

func main() {
	var c Config
	err := envconfig.Process("fake", &c)

	log.Printf("env { FAKE_HOST=%s, FAKE_PORT=%s, FAKE_PROXY_MEMCACHED_HOST=%s,"+
		" FAKE_PROXY_MEMCACHED_PORT=%s, FAKE_CLUSTER_NODES=%s, FAKE_TRACE=%t }",
		c.Host, c.Port, c.ProxyMemcachedHost, c.ProxyMemcachedPort, c.ClusterNodes, c.Trace)

	if err != nil {
		log.Fatal(err.Error())
	}

	server := New(&c)

	server.Run()
}
