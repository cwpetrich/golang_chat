package main

import (
	"bufio"
	"fmt"
	"log"
	//"net"
	//"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"
)

const defaultMachine = "localhost"
const defaultPort = "3410"

type Nothing struct{}

type SayRecord struct {
	User    string
	Message string
}

type TellRequest struct {
	User    string
	Target  string
	Message string
}

func List(client *rpc.Client, nothing Nothing) {
	users := make([]string, 0)
	err := client.Call("Server.List", nothing, &users)
	if err != nil {
		log.Fatal("List:", err)
	}
	fmt.Println("Users currently logged in")
	for i := range users {
		fmt.Println(users[i])
	}
}


func ParseMessage(input string, user string, client *rpc.Client) error {
	nothing := Nothing{}
	original_input := input
	stripped_input := strings.TrimSpace(input)
	stripped_input = strings.ToLower(stripped_input)
	type reply int
	switch {

	case strings.HasPrefix(stripped_input, "tell"):
		original_input_split := strings.SplitN(original_input, " ", 3)
		original_message := original_input_split[2]
		tell_list := strings.SplitN(stripped_input, " ", 3)
		tellRequest := TellRequest{user, tell_list[1], original_message}
		err := client.Call("Server.Tell", tellRequest, &nothing)
		if err != nil {
			log.Fatal("Tell:", err)
		}

	case strings.HasPrefix(stripped_input, "say"):
		original_input_split := strings.SplitN(original_input, " ", 2)
		original_message := original_input_split[1]
		sayRecord := SayRecord{user, original_message}
		err := client.Call("Server.Say", sayRecord, &nothing)
		if err != nil {
			log.Fatal("Say:", err)
		}

	case strings.HasPrefix(stripped_input, "list"):
		List(client, nothing)
	case strings.HasPrefix(stripped_input, "logout"):
		err := client.Call("Server.Logout", user, &nothing)
		if err != nil {
			log.Fatal("Logout:", err)
		}
		client.Close()
		

	case strings.HasPrefix(stripped_input, "shutdown"):
		err := client.Call("Server.Shutdown", nothing, &nothing)
		if err != nil {
			log.Fatal("Shutdown:", err)
		}

	default:
		fmt.Println("Valid commands are:")
		fmt.Println("tell :: (followed by a username you wish to send a message to (ex. tell cpetrich <message>)")
		fmt.Println("say,")
		fmt.Println("list,")
		fmt.Println("logout,")
		fmt.Println("shutdown")
	}
	return nil
}

func main() {
	nothing := Nothing{}
	username := ""
	machine := ""
	port := ""

	// If only username given, dafault host and port are used
	if len(os.Args) == 2 {
		username = os.Args[1]
		machine = defaultMachine
		port = defaultPort
	// If username and port given, default host is used
	} else if len(os.Args) == 3 {
		username = os.Args[1]
		address_list := strings.SplitN(os.Args[2], ":", 2)
		if len(address_list) == 1 {
			machine = address_list[0]
			port = defaultPort
		}else if address_list[0] != "" && address_list[0] != "" {
			machine = address_list[0]
			port = address_list[1]
		}else if address_list[0] != "" && address_list[1] == "" {
			machine = address_list[0]
			port = defaultPort
		}else if address_list[0] == "" && address_list[1] != "" {
			machine = defaultMachine
			port = address_list[1]
		}else{
			machine = defaultMachine
			port = defaultPort
		}
	// No defaults are used
	//} else if len(os.Args) == 4 {
	//	username = os.Args[1]
	//	machine = os.Args[3]
	//	port = os.Args[2]
	// Displays valid arguement options for ./chatclient 
	} else {
		fmt.Println("Valid Arguements:\nos.Args[0]=./chatclient\nos.Args[1]=<username>\nos.Args[2]=<port_number> (OPTIONAL, Defaults to 3410)\nos.Args[3]=<machine_name> (OPTIONAL, Defaults to localhost)")
		os.Exit(1)
	}

	// Forming connection to server at hostname = machine and port = port.
	client, err := rpc.DialHTTP("tcp", machine+":"+port)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// Calls Register on server, performs login for user to server
	err = client.Call("Server.Register", username, &nothing)
	if err != nil {
		log.Fatal("Register:", err)
	}

	// Calls List on server, Returns list of users currently logged in
	List(client, nothing)
	scanner := bufio.NewScanner(os.Stdin)
	messages := make([]string, 0)
	
	go func() {
		for {
			// Calls CheckMessages on server, prints messages to terminal from the user's queue.
			err = client.Call("Server.CheckMessages", username, &messages)
			if err != nil {
				log.Fatal("CheckMessages:", err)
			}
			for i := range messages {
				fmt.Println(messages[i])
			}
			time.Sleep(time.Second)
		}
	}()

	for scanner.Scan(){
		text := scanner.Text()
		err = ParseMessage(text, username, client)
		if err != nil {
			log.Fatal("ParseMessage:", err)
		}
	}
}
