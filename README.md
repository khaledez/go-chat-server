# Telnet Chat Server

## Requirements
- Go version 1.13
- Telnet client

## Steps to run it locally
```bash
# clone this repository
# then ..
$ cd go-chat-server
$ go run .
```

## Configuration

By default, the server binds to port `2323` and writes log file into current working directory.

To change the default configuration, create a file named `config.yml` in the current working directory, with the following options:
```yml
port: 2323
log_file_path: ~/Downloads
```

## Design decisions
The main component is the room (defined in `chat/room.go`), a chat room has a name, log file and a list of online users.

A user (defined in `chat/user.go`) has a name and a channel

A channel (defined in `chat/channel.go`) is an abstraction to a communication medium that transfer strings as messages between the server and the client. The server listens to messages from the channel, and sends messages to the channel.

Every user in the chat room has it's own go routine.

### Chat commands
A user can type `/help` in the chat prompt to get the list of available commands.

Currently supported commands: `/exit` and `/online`, which prints the number of online users in the room