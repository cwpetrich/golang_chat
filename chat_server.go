package main

import (
	//"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"os"
	//"strings"
)

const defaultPort = "3410"

type Nothing struct{}

type Messages struct {
	Users map[string][]string
	End   chan bool
}

type TellRequest struct {
	User    string
	Target  string
	Message string
}

type SayRecord struct {
	User    string
	Message string
}

type Server chan *Messages

func (s Server) Register(user string, reply *Nothing) error {
	msg := fmt.Sprintf("*** %s has logged in", user)
	log.Print(msg)
	elt := <-s
	for target, queue := range elt.Users {
		elt.Users[target] = append(queue, msg)
	}
	elt.Users[user] = nil
	s <- elt
	return nil
}

func (s Server) List(request *Nothing, users *[]string) error {
	elt := <-s
	for target, _ := range elt.Users {
		*users = append(*users, target)
	}
	s <- elt
	sort.Strings(*users)
	return nil
}


func (s Server) Tell(tellRequest TellRequest, reply *Nothing) error {
	msg := fmt.Sprintf("(Private Message)-> %s: %s", tellRequest.User, tellRequest.Message)
	log.Print(msg)
	elt := <-s
	elt.Users[tellRequest.Target] = append(elt.Users[tellRequest.Target], msg)
	s <- elt
	return nil
}


func (s Server) Say(sayRecord SayRecord, reply *Nothing) error {
	msg := fmt.Sprintf(sayRecord.User+": "+sayRecord.Message)
	elt := <-s
	for target, queue := range elt.Users {
		elt.Users[target] = append(queue, msg)
	}
	//elt.Users[sayRecord.user] = nil
	s <- elt
	return nil
}

func (s Server) Logout(user string, reply *Nothing) error {
	msg := fmt.Sprintf("*** %s has logged out", user)
	log.Print(msg)
	elt := <-s
	for target, queue := range elt.Users {
		elt.Users[target] = append(queue, msg)
	}
	delete(elt.Users, user)
	s <- elt
	return nil
}

func (s Server) CheckMessages(user string, messages *[]string) error {
	elt := <-s
	if queue, present := elt.Users[user]; present {
		*messages = queue
		elt.Users[user] = nil
	} else {
		*messages = []string{"You are not logged in, " + user}
	}
	s <- elt
	return nil
}

func (s Server) Shutdown(request *Nothing, reply *Nothing) error {
	log.Printf("***The server is shutting down...")
	elt := <-s
	elt.End <- true
	return nil
}

func main() {
	port := ""
	if len(os.Args) == 2 {
		port = os.Args[1]
	}else if len(os.Args) == 1 {
		port = defaultPort
	}else {
		fmt.Println("Valid Arguements:\nos.Args[0]=./chatserver\nos.Args[1]=<port_number> (OPTIONAL, Defaults to 3410)")
	}
	//flag.StringVar(&port, "port", defaultPort, "Specify port number for server to listen on")
	//flag.Parse()
	address := net.JoinHostPort("", port)
	log.Println()

	elt := &Messages{
		Users: make(map[string][]string),
		End:   make(chan bool),
	}
	server := Server(make(chan *Messages, 1))
	server <- elt
	rpc.Register(server)
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal ("listen error: ", err)
	}
	log.Println("Listening on ", address)
	log.Println()
	go http.Serve(listener, nil)
	<-elt.End
}
