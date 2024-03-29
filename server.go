package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
)

type handle func(http.ResponseWriter, *http.Request) error

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	select {
	case err = <-errChan:
	default:
		err = h(w, r)
	}

	if err != nil {
		log.Print(err)
	}
}

func main() {
	server()
}

//go:embed public/index.html
var html []byte

//go:embed public/styles.css
var styles []byte

//go:embed public/scripts.js
var scripts []byte

//go:embed public/game/styles.css
var gameStyles []byte

//go:embed public/game/scripts.js
var gameScripts []byte

//go:embed public/game/index.html
var gameHTML []byte

// Removed from main func for testability
func server() {
	go gameManager()

	http.Handle("/", handle(func(w http.ResponseWriter, r *http.Request) error {
		return fileHandler(html, "text/html", w)
	}))

	http.Handle("/public/styles.css", handle(func(w http.ResponseWriter, r *http.Request) error {
		return fileHandler(styles, "text/css", w)
	}))

	http.Handle("/public/scripts.js", handle(func(w http.ResponseWriter, r *http.Request) error {
		return fileHandler(scripts, "text/javascript", w)
	}))

	http.Handle("/public/game/styles.css", handle(func(w http.ResponseWriter, r *http.Request) error {

		return fileHandler(gameStyles, "text/css", w)
	}))

	http.Handle("/public/game/scripts.js", handle(func(w http.ResponseWriter, r *http.Request) error {
		return fileHandler(gameScripts, "text/javascript", w)
	}))

	http.Handle("/game/", handle(func(w http.ResponseWriter, r *http.Request) error {
		return fileHandler(gameHTML, "text/html", w)
	}))

	http.Handle("/ws/", handle(func(w http.ResponseWriter, r *http.Request) error {
		ws, err := handshake(w, r)
		// As per RFC6455:
		// If the invalid data is sent during the WebSocket handshake,
		// the server SHOULD return an appropriate HTTP [RFC2616] status code.
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return fmt.Errorf("error with ws handshake: %w", err)
		}

		err = WsHandler(w, r, ws)

		// TODO: Change error handling because may no longer use HTTP
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		return nil
	}))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func fileHandler(file []byte, contentType string, w http.ResponseWriter) error {
	w.Header().Add("Content-type", contentType)
	_, err := w.Write(file)
	if err != nil {
		return fmt.Errorf("error sending response: %w", err)
	}

	return nil
}

var errChan = make(chan error)

type ServerMessage struct {
	// Forward is a message sent from a client that's being forwarded to another
	Forward map[string]interface{} `json:",omitempty"`
	// Players is to alert who is already connected. Use to check if a given
	// player is the first one.
	Players []Author `json:",omitempty"`
	// Send all the lines at the end of games
	TheEnd []Line `json:",omitempty"`
	// LineAuthors is used to send the authors of each story line to a new player
	LineAuthors []Author `json:",omitempty"`
}

type ClientMessage struct {
	// NewPlayer serves for a new player to send its name
	NewPlayer Author `json:",omitempty"`
	// Broadcast is an arbitrary message from the server used for joining, leaving,
	// status updates (who wrote which line).
	// The client sends arbitrary JSON as string.
	// The server shouldn't be looking inside this message.
	Broadcast map[string]interface{} `json:",omitempty"`
	// NextPlayer is the latest game line sent to the next player
	NextPlayer string `json:",omitempty"`
}

type MessageChan struct {
	Message ClientMessage
	Client  client
}

type Player struct {
	Client   client
	UserName Author
}

type client chan ServerMessage
type gameChannels struct {
	messages chan MessageChan
	leaving  chan client
}
type replyChan chan gameChannels
type gameSession struct {
	uuid string
	replyChan
}

var createGame = make(chan gameSession)

// func handleNewPlayer() {

// }

func convertPlayerOrderToString(po []Player) []Author {
	userNames := make([]Author, len(po))
	for i, p := range po {
		userNames[i] = p.UserName
	}

	return userNames
}

func convertLinesToAuthors(lines []Line) []Author {
	authors := make([]Author, len(lines))
	for i, l := range lines {
		authors[i] = l.Author
	}
	return authors
}

type Author string

type Line struct {
	Text string
	Author
}

// gameManager controls access to the mapping from UUID to game channels
func gameManager() {
	games := make(map[string]gameChannels)
	for {
		gameSess := <-createGame
		uuid := gameSess.uuid
		replyCh := gameSess.replyChan
		_, ok := games[uuid]
		if !ok {
			gameCh := gameChannels{
				messages: make(chan MessageChan),
				leaving:  make(chan client),
			}
			games[uuid] = gameCh
			go broadcast(gameCh)
		}
		replyCh <- games[uuid]
	}
}

func broadcast(gameCh gameChannels) {
	var playerOrder []Player
	playerTurn := 0
	var lines []Line
	isTheEnd := false
	for {
		select {
		case msg := <-gameCh.messages:
			switch {
			case msg.Message.NewPlayer != "":
				fmt.Printf("msg from server <3 NewPlayer: %+v\n", msg)
				// TODO: fxn call...
				playerOrder = append(playerOrder, Player{Client: msg.Client, UserName: msg.Message.NewPlayer})
				for _, cli := range playerOrder {
					cli.Client <- ServerMessage{Players: convertPlayerOrderToString(playerOrder)}
				}
				lenLines := len(lines)
				if isTheEnd {
					msg.Client <- ServerMessage{TheEnd: lines}
				} else if lenLines > 0 {
					msg.Client <- ServerMessage{LineAuthors: convertLinesToAuthors(lines)}
				}
			case msg.Message.Broadcast != nil:
				fmt.Printf("msg from server <3 Broadcast: %+v\n", msg)
				// TODO: fxn call...
				for _, cli := range playerOrder {
					cli.Client <- ServerMessage{Forward: msg.Message.Broadcast}
				}
			case msg.Message.NextPlayer != "":
				lines = append(lines, Line{
					Author: playerOrder[playerTurn].UserName,
					Text:   msg.Message.NextPlayer,
				})
				fmt.Printf("msg from server <3 NextPlayer: %+v\n", msg)
				if playerTurn == len(playerOrder)-1 {
					playerTurn = 0
				} else {
					playerTurn++
				}
				// TODO: fxn call...
				c := playerOrder[playerTurn]
				for _, cli := range playerOrder {
					// End game automatically
					if len(lines) >= 10 && len(lines) >= len(playerOrder)*2 {
						isTheEnd = true
						cli.Client <- ServerMessage{TheEnd: lines}
					} else if c.Client == cli.Client {
						c.Client <- ServerMessage{Forward: map[string]interface{}{
							"NextPlayer": msg.Message.NextPlayer,
						}}

					}
				}
			default:
				log.Print("msg defaul <3 ", msg)
			}

			// log.Print("msg from server <3 ", msg)

			fmt.Printf("playerOrder: %+v\n", playerOrder)
			log.Println(playerTurn)
		case cli := <-gameCh.leaving:
			for i, c := range playerOrder {
				if c.Client == cli {
					playerOrder = append(playerOrder[:i], playerOrder[i+1:]...)
					if len(playerOrder) == 1 {
						playerTurn = 0
					} else if playerTurn == 0 {
						playerTurn = len(playerOrder) - 1
					} else {
						playerTurn--
					}
				}
				c.Client <- ServerMessage{Players: convertPlayerOrderToString(playerOrder)}
			}
			close(cli)
		}
	}
}

func WsHandler(w http.ResponseWriter, req *http.Request, ws *Ws) error {
	defer ws.Close()

	// Frame
	// A long message to test endianness
	// err := ws.Send("hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 hello, this is Lila's ws server <3 <3 <3")
	// if err != nil {
	// 	return err
	// }
	// not frame
	// ws.write([]byte("hello websocket <3"))

	ch := make(client)
	go clientWriter(ws, ch)
	u := path.Base(req.URL.Path)
	replyCh := make(replyChan)
	createGame <- gameSession{
		uuid:      u,
		replyChan: replyCh,
	}
	gameChans := <-replyCh

loop:
	for {
		msg, opcode, err := ws.Recv()
		if err != nil {
			switch err {
			case io.EOF:
				log.Println("end of file error: ", err.Error())
				break loop
			default:
				log.Println()
				log.Println("closing error: ", err.Error())
				break loop
			}
		}

		if opcode == 0x9 || msg == "PING" {
			// If receive PING, send PONG back
			// if the connection wasn't closed,
			// TODO: sending back the same Application
			// Data from the PING
			ws.Pong()
			continue
		}

		// Make sure to broadcast only text messages,
		// not Control frames like Close, Ping, and Pong
		if opcode == 1 && msg != "PING" && msg != "PONG" {
			log.Println(" WOOHOO msg from browser <3 ", msg)
			// Send MessageChan that contains the client channel
			var m ClientMessage
			err := json.Unmarshal([]byte(msg), &m)
			if err != nil {
				return fmt.Errorf("unmarshall error: %w", err)
			}
			fmt.Printf("CLientMessage: %+v\n", m)
			gameChans.messages <- MessageChan{Client: ch, Message: m}
		}
	}

	gameChans.leaving <- ch
	gameChans.messages <- MessageChan{Client: ch}
	return nil
}

func clientWriter(ws *Ws, ch client) {
	for msg := range ch {
		m, err := json.Marshal(msg)
		if err != nil {
			log.Println(err.Error())
			ws.Close()
		}
		err = ws.SendMsg(string(m))
		if err != nil {
			log.Println(err)
			ws.Close()
		}
	}
}

func handshake(w http.ResponseWriter, r *http.Request) (*Ws, error) {
	httpStatus, err := validateWsRequest(r)
	if err != nil {
		http.Error(w, err.Error(), httpStatus)
		return nil, err
	}
	// As per RFC6455:
	// For this header field [Sec-WebSocket-Key], the server has to take the value (as present
	// in the header field, e.g., the base64-encoded [RFC4648] version minus
	// any leading and trailing whitespace) and concatenate this with the
	// Globally Unique Identifier (GUID, [RFC4122]) "258EAFA5-E914-47DA-
	// 95CA-C5AB0DC85B11" in string form, which is unlikely to be used by
	// network endpoints that do not understand the WebSocket Protocol.  A
	// SHA-1 hash (160 bits) [FIPS.180-3], base64-encoded (see Section 4 of
	// [RFC4648]), of this concatenation is then returned in the server's
	// handshake.
	wsKeyConcat := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key")) + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	wsBytes := []byte(wsKeyConcat)
	hasher := sha1.New()
	hasher.Write(wsBytes)
	sha := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	w.Header().Add("Upgrade", "websocket")
	w.Header().Add("Connection", "Upgrade")
	w.Header().Add("Sec-WebSocket-Accept", sha)
	w.Header().Add("Sec-WebSocket-Protocol", "chat")
	w.WriteHeader(101)

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("webserver doesn't support http hijacking")
	}
	conn, bufwr, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	ws := &Ws{conn, bufwr, r}
	return ws, nil
}

func validateWsRequest(r *http.Request) (int, error) {
	// An HTTP/1.1 or higher GET request, including a "Request-URI"
	// [RFC2616] that should be interpreted as a /resource name/
	// defined in Section 3 (or an absolute HTTP/HTTPS URI containing
	// the /resource name/).
	method := r.Method
	if method != "GET" {
		return http.StatusForbidden, fmt.Errorf("forbidden method %s\n, must be 'GET'", method)
	}

	if !r.ProtoAtLeast(1, 1) {
		return http.StatusForbidden, fmt.Errorf("forbidden HTTP protocol version %s\n, must be 1.1 or higher", r.Proto)
	}

	// As per RFC6455:
	// The client includes the hostname in the |Host| header field of its
	// handshake as per [RFC2616], so that both the client and the server
	// can verify that they agree on which host is in use.
	host := r.Host
	// fmt.Printf("host: %s\n", host)
	if host != "localhost:8080" {
		return http.StatusForbidden, fmt.Errorf("forbidden host: %s", host)
	}

	// An |Upgrade| header field containing the value "websocket",
	// treated as an ASCII case-insensitive value.
	upgrade := r.Header.Get("Upgrade")
	if upgrade != "websocket" {
		return http.StatusBadRequest, fmt.Errorf("invalid Upgrade header %s, must be 'websocket'", upgrade)
	}

	wsKey := r.Header.Get("Sec-WebSocket-Key")

	// As per RFC6455:
	// The request MUST include a header field with the name
	// |Sec-WebSocket-Key|.  The value of this header field MUST be a
	// nonce consisting of a randomly selected 16-byte value that has
	// been base64-encoded (see Section 4 of [RFC4648]).  The nonce
	// MUST be selected randomly for each connection.
	wsKeyBytes, err := base64.StdEncoding.DecodeString(wsKey)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error decoding Sec-WebSocket-Key header: %s", err)
	}
	if len(wsKeyBytes) != 16 {
		return http.StatusBadRequest, fmt.Errorf("invalid Sec-WebSocket-Key header length of %d, must be 16-bytes long", len(wsKeyBytes))
	}

	return 0, nil
}

func (ws *Ws) Close() {
	// Send closing frame
	err := ws.Send("", 0x8)
	if err != nil {
		errChan <- err
	}

	err = ws.conn.Close()
	if err != nil {
		errChan <- fmt.Errorf("error closing ws connection: %w", err)
	}
}

func (ws *Ws) SendMsg(msg string) error {
	err := ws.Send(msg, 1)
	return err
}

func (ws *Ws) write(data []byte) error {
	_, err := ws.bufrw.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to ws connection: %w", err)
	}
	err = ws.bufrw.Flush()
	if err != nil {
		return fmt.Errorf("error flushing ws writer: %w", err)
	}

	return nil
}

func (ws *Ws) Recv() (string, uint8, error) {
	// TODO: opcode, fail if RSV values are not 0,
	// fail if not masked, unmask
	head := make([]byte, 2)
	_, err := io.ReadFull(ws.bufrw, head)
	if err != nil {
		return "", 0, fmt.Errorf("error reading frame head: %w", err)
	}
	parsedFrame := parseFrameHead(head)
	// TODO: validate rsv1, rsv2, rsv3, and mask
	if parsedFrame.payLen == 126 {
		var extPayLen16 uint16
		// As per RFC6455:
		// in network order (big endian)
		err := binary.Read(ws.bufrw, binary.BigEndian, &extPayLen16)
		if err != nil {
			return "", 0, fmt.Errorf("error reading extended pay length (16): %w", err)
		}
		parsedFrame.extPayLen1 = extPayLen16
	}
	if parsedFrame.payLen == 127 {
		var extPayLen64 uint64
		err := binary.Read(ws.bufrw, binary.BigEndian, &extPayLen64)
		if err != nil {
			return "", 0, fmt.Errorf("error reading extended pay length (64): %w", err)
		}
		parsedFrame.extPayLen2 = extPayLen64
	}
	maskKey := make([]byte, 4)
	_, err = io.ReadFull(ws.bufrw, maskKey)
	if err != nil {
		return "", 0, fmt.Errorf("error reading masking key: %w", err)
	}

	parsedFrame.maskKey = maskKey

	payLen := getPayloadLength(parsedFrame)
	pay := make([]byte, payLen)
	_, err = io.ReadFull(ws.bufrw, pay)
	if err != nil {
		return "", 0, fmt.Errorf("error reading payload: %w", err)
	}
	parsedFrame.payload = pay
	unmasked := unmaskPayload(parsedFrame)
	return unmasked, parsedFrame.opcode, nil
}

func parseFrameHead(frame []byte) Frame {
	parsedFrame := Frame{}

	var byte1 uint8 = frame[0]
	parsedFrame.fin = byte1 >> 7
	parsedFrame.rsv1 = ((1 << 6) & byte1) >> 6
	parsedFrame.rsv2 = ((1 << 5) & byte1) >> 5
	parsedFrame.rsv3 = ((1 << 4) & byte1) >> 4
	parsedFrame.opcode = 0x01 & byte1

	var byte2 uint8 = frame[1]
	parsedFrame.mask = byte2 >> 7
	parsedFrame.payLen = 0x7f & byte2

	return parsedFrame
}

func getPayloadLength(parsedFrame Frame) int {
	if parsedFrame.payLen < 126 {
		return int(parsedFrame.payLen)
	}

	if parsedFrame.payLen == 126 {
		return int(parsedFrame.extPayLen1)
	}

	return int(parsedFrame.extPayLen2)
}

func parseFramePayload(frame []byte, parsedFrame Frame, idx int) {
	parsedFrame.payload = frame[idx:]
}

func unmaskPayload(frame Frame) string {
	key := frame.maskKey
	payLen := getPayloadLength(frame)
	unmasked := make([]byte, payLen)
	keyIdx := 0
	for i, e := range frame.payload {

		unmasked[i] = e ^ key[keyIdx]
		if keyIdx == len(key)-1 {
			keyIdx = 0
			continue
		}
		keyIdx++
	}

	unmaskedStr := string(unmasked)
	return unmaskedStr
}

// func validateAndReturnFrame(frame Frame) error {

// }

func (ws *Ws) Pong() {
	// TODO: send same application data back from PING
	ws.Send("PONG", 0xA)
}

func (ws *Ws) Send(msg string, opcd uint8) error {
	pay := []byte(msg)
	payLen := len(pay)

	var fin uint8 = 1
	var rsv1 uint8 = 0
	var rsv2 uint8 = 0
	var rsv3 uint8 = 0
	var opcode uint8 = opcd
	var masked uint8 = 0

	frame := new(bytes.Buffer)
	byte1 := (fin << 7) | (rsv1 << 6) | (rsv2 << 5) | (rsv3 << 4) | opcode
	// log.Println(byte1)
	err := frame.WriteByte(byte1)
	if err != nil {
		return fmt.Errorf("error writing byte 1 to ws frame: %w", err)
	}

	switch {
	case payLen < 126:
		// 7 bits to denote payload length (in bytes)
		byte2 := (masked << 7) | uint8(payLen)
		err = frame.WriteByte(byte2)
		if err != nil {
			return fmt.Errorf("error writing byte 2 to ws frame (<126): %w", err)
		}
	case payLen < (1 << 16):
		// 7+16 bits
		byte2 := (masked << 7) | (uint8(126))
		err = frame.WriteByte(byte2)
		if err != nil {
			return fmt.Errorf("error writing byte 2 to ws frame (==126): %w", err)
		}
		bytes34 := uint16(payLen)
		err = binary.Write(frame, binary.BigEndian, bytes34)
		if err != nil {
			return fmt.Errorf("error writing extended payload length to ws frame (16): %w", err)
		}
	default:
		// 7+64 bits
		byte2 := (masked << 7) | (uint8(127))
		err = frame.WriteByte(byte2)
		if err != nil {
			return fmt.Errorf("error writing byte 2 to ws frame (==127): %w", err)
		}
		nextBytes := uint64(payLen)
		err = binary.Write(frame, binary.BigEndian, nextBytes)
		if err != nil {
			return fmt.Errorf("error writing extended payload length to ws frame (64): %w", err)
		}
	}
	_, err = frame.Write(pay)
	if err != nil {
		return fmt.Errorf("error writing payload to ws frame: %w", err)
	}
	ws.write(frame.Bytes())
	return nil
}

type Ws struct {
	conn    net.Conn
	bufrw   *bufio.ReadWriter
	request *http.Request
}

type Frame struct {
	fin        uint8
	rsv1       uint8
	rsv2       uint8
	rsv3       uint8
	opcode     uint8
	mask       uint8
	payLen     uint8
	extPayLen1 uint16
	extPayLen2 uint64
	maskKey    []byte
	payload    []byte
}
