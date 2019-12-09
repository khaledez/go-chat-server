package main

import (
	"chat-server/chat"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"os/signal"
)

const (
	configPort    = "port"
	configLogPath = "log_file_path"
)

func main() {
	go watchForSignals()

	// I'm using Viper for config management, it will look for a file called "config.yml"
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetDefault(configPort, "2323")
	viper.SetDefault(configLogPath, ".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file (config.yml) not found, using the default configuration")
		} else {
			log.Printf("Error reading config file, falling back to default values")
		}
	}

	port := viper.GetInt(configPort)
	logFilePath := viper.GetString(configLogPath)

	r, err := chat.NewRoom("Torbit Chat Server", logFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(listenAndServe(port, func(conn net.Conn) {
		r.Join(chat.NewChannel(conn), conn.RemoteAddr())
	}))
}

func watchForSignals() {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Kill, os.Interrupt)

	<-ch
	fmt.Println("\nGoodbye!")
	os.Exit(0)
}

func listenAndServe(port int, handler func(net.Conn)) error {
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return NewServerError("Error starting server: %v", err)
	}
	defer server.Close()
	log.Printf("Listening on port %d\n", port)

	for {
		conn, err := server.Accept()
		if err != nil {
			return NewServerError("Error accepting connection: %v", err)
		}

		go handler(conn)
	}
}

// ServerError to throw
type ServerError struct{ msg string }

// NewServerError creates a ServerError with formated error mesage
func NewServerError(msg string, cause error) *ServerError {
	return &ServerError{fmt.Sprintf(msg, cause)}
}

func (e *ServerError) Error() string {
	return e.msg
}
