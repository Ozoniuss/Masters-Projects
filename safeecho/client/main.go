package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	crypt "safechat/encryption"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "6699"
	SERVER_TYPE = "tcp"
)

const (
	NO_HEADER    byte = 255
	CLIENT_HELLO byte = 0
	SERVER_HELLO byte = 1
	CLIENT_DONE  byte = 2
	SERVER_DONE  byte = 3
	ERROR        byte = 4
	CLIENT_MSG   byte = 5
	SERVER_MSG   byte = 6
	CLIENT_CLOSE byte = 7
	SERVER_CLOSE byte = 8
)

type ConnState struct {
	pubKey *crypt.PublicKey
	symKey *[32]byte
}

func newState() ConnState {
	return ConnState{
		pubKey: nil,
		symKey: nil,
	}
}

func readMessage() (byte, string) {
	var msg string

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		msg = scanner.Text()
	}

	headerDelimPos := strings.Index(msg, ":")
	if headerDelimPos == -1 {
		return CLIENT_MSG, msg
	} else {
		header, err := strconv.ParseUint(msg[:headerDelimPos], 10, 8)
		if err != nil {
			return CLIENT_MSG, msg
		}
		return byte(header), msg[headerDelimPos+1:]
	}
}

func writeMsg(typ byte, msg string, s *ConnState) []byte {
	sends := []byte{typ}
	if s.symKey != nil && msg != "" {
		msg = fmt.Sprintf("%s", crypt.EncryptAES(s.symKey[:], []byte(msg)))
	}
	sends = append(sends, []byte(msg)...)
	return sends
}

func main() {
	//establish connection
	connection, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		panic(err)
	}

	state := newState()

	autoConnect(connection, &state)
	//processMessage(connection, &state)

	for {
		typ, msg := readMessage()
		sends := writeMsg(typ, msg, &state)
		_, err := connection.Write(sends)
		if err != nil {
			panic(err)
		}
		// send the hello automatically so it's taken care of by the client
		displayMessage(connection, &state)
	}
}

func autoConnect(connection net.Conn, s *ConnState) {
	sends := []byte{CLIENT_HELLO}
	connection.Write(sends)

	// Receives server hello
	buffer, mLen, err := readFromServer(connection)
	if err != nil {
		fmt.Printf("an error occured: %v", err)
	}
	header := buffer[0]
	content := buffer[1:mLen]

	if header != SERVER_HELLO {
		fmt.Println("an error occured during the handshake")
		os.Exit(1)
	}

	// Generate symmetric key after client hello
	pubKey := &crypt.PublicKey{}
	pubKey.Unmarshal(content)
	s.pubKey = pubKey

	symKey := generateSymKey()
	fmt.Printf("[server hello] generated sym key: %v\n", symKey)

	msg := pubKey.EncryptString(symKey[:])
	connection.Write(writeMsg(CLIENT_DONE, msg, s))

	s.symKey = &symKey

	// Receives server done
	buffer, mLen, err = readFromServer(connection)
	if err != nil {
		fmt.Printf("an error occured: %v", err)
	}
	header = buffer[0]
	if header != SERVER_DONE {
		fmt.Println("did not receive server done")
		os.Exit(1)
	}
	fmt.Println("[server done] handshake complete")
}

func readFromServer(connection net.Conn) ([]byte, int, error) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		return nil, 0, err
	}
	if mLen == 0 {
		return nil, 0, errors.New("Received null message")
	}
	return buffer, mLen, nil
}

func displayMessage(connection net.Conn, s *ConnState) (byte, error) {

	buffer, mLen, err := readFromServer(connection)
	if err != nil {
		fmt.Printf("an error occured: %v", err)
	}
	header := buffer[0]
	content := buffer[1:mLen]

	switch header {
	case SERVER_HELLO:

		fmt.Println("[server hello] received server hello")
		pubKey := &crypt.PublicKey{}
		pubKey.Unmarshal(content)
		fmt.Printf("[server hello] public key is %+v\n", pubKey)

	case SERVER_MSG:
		fmt.Printf("[message] server encrypted message as: %s\n", base64.URLEncoding.EncodeToString(content))

	case SERVER_DONE:
		fmt.Println("[server done] handshake complete")

	case ERROR:
		fmt.Printf("[error] received error: %s\n", content)

	default:
		fmt.Println("[error] handshake complete")
	}
	return NO_HEADER, nil
}

func generateSymKey() [32]byte {
	var key32 [32]byte
	key := make([]byte, 32)
	rand.Read(key)
	copy(key32[:], key[:])
	return key32
}
