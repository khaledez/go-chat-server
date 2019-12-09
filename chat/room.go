package chat

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

// Room is chat room
type Room struct {
	sync.Mutex
	name        string
	logFile     *os.File
	onlineUsers []User
}

// NewRoom creates a new chat room
func NewRoom(name, logFilePath string) (*Room, error) {

	p, err := expandPath(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error while getting log file path: %w", err)
	}

	logFileName := path.Join(p, strings.ReplaceAll(name, " ", "_")+".dat")
	log.Printf("Log file location: %s\n", logFileName)

	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error opening/creating log file: %v", err)
	}

	return &Room{name: name, logFile: logFile}, nil
}

// Join adds a user into the room
func (r *Room) Join(ch *Channel, ipAddr net.Addr) error {
	ch.Write(fmt.Sprintf("Welcome to %s chat.%sWhat's your name: ", r.name, crlf))

	name, err := ch.ReadString()
	if err != nil {
		return err
	}
	log.Printf("%q joined from %s!", name, ipAddr.String())
	user := User{Name: name, Channel: ch}

	r.newUser(&user)
	r.sendUpdate(fmt.Sprintf("%s joined!", name))

	for {
		ch.Write(r.prompt())

		input, err := ch.ReadString()

		if err != nil {
			r.logout(name, ch)
			return err
		}

		writeBytes(ch.target, []byte("\x1B[1A\x1B[2K\x1B[1G"))

		if len(input) == 0 {
			continue
		} else if input[0] == '/' {
			handleCommand(r, input[1:], &user)
		} else {
			r.send(name, input)
		}
	}
}

const helpString = `Commands:
/exit - log out 
/online - shows how many users online` + crlf

func handleCommand(r *Room, c string, user *User) {
	// command
	switch c {
	case "help":
		user.Channel.Write(helpString)
	case "exit":
		r.logout(user.Name, user.Channel)
		break
	case "online":
		user.Channel.Write(fmt.Sprintf("%d users online%s", len(r.onlineUsers), crlf))
	default:
		user.Channel.Write("ERROR: unknown command" + crlf)
	}
}

func (r *Room) prompt() string {
	return fmt.Sprintf("#%s> ", r.name)
}

func (r *Room) newUser(u *User) {
	r.Mutex.Lock()
	r.onlineUsers = append(r.onlineUsers, *u)
	r.Mutex.Unlock()
}

func (r *Room) removeUser(name string) {
	i := sort.Search(len(r.onlineUsers), func(i int) bool { return r.onlineUsers[i].Name == name })

	if i >= 0 {
		// the below code is slightly modified version of https://github.com/golang/go/wiki/SliceTricks#delete
		size := len(r.onlineUsers)
		if i < size-1 {
			copy(r.onlineUsers[i:], r.onlineUsers[i+1:])
		}
		r.onlineUsers[size-1] = User{} // zero-value of user
		r.onlineUsers = r.onlineUsers[:size-1]
	}
}

func (r *Room) send(from, msg string) {
	formatted := fmt.Sprintf("%s - %s >> %s", time.Now().Format(timeFormat), from, msg+crlf)
	r.broadcast(formatted)
}

func (r *Room) sendUpdate(msg string) {
	formatted := fmt.Sprintf("%s = %s", time.Now().Format(timeFormat), msg+crlf)
	r.broadcast(formatted)
}

func (r *Room) broadcast(msg string) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(r.onlineUsers))

	for i := range r.onlineUsers {
		go func(wg *sync.WaitGroup, u *User) {
			u.Channel.Write(msg)
			u.Channel.Write(r.prompt())
			wg.Done()
		}(&waitGroup, &r.onlineUsers[i])
	}
	waitGroup.Wait()

	// log the message
	r.Lock()
	if _, err := r.logFile.WriteString(msg); err != nil {
		log.Printf("Error writing to log file: %v\n", err)
	}
	if err := r.logFile.Sync(); err != nil {
		log.Printf("Error flushing data to log file: %v\n", err)
	}
	r.Unlock()
}

func (r *Room) logout(name string, ch *Channel) {
	r.Lock()

	ch.Close()
	r.removeUser(name)

	r.Unlock()

	r.sendUpdate(fmt.Sprintf("%s left!", name))
}
