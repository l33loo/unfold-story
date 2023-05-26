package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	go broadcast()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/html")
		f, err := os.Open("./public/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/css")
		f, err := os.Open("./public/styles.css")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/scripts.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/javascript")
		f, err := os.Open("./public/scripts.js")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/game/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/css")
		f, err := os.Open("./public/game/styles.css")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/game/scripts.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/javascript")
		f, err := os.Open("./public/game/scripts.js")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./public/game/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := handshake(w, r)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = WsHandler(w, r, ws)
		// TODO: Change error handling because may no longer use HTTP
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Message struct {
	Join      string `json:",omitempty"`
	Entering  string `json:",omitempty"`
	Leaving   string `json:",omitempty"`
	Line      string `json:",omitempty"`
	EmptyLine bool   `json:",omitempty"`
	Start     bool   `json:",omitempty"`
}

type client chan<- string

var (
	entering = make(chan client)
	messages = make(chan string)
	leaving  = make(chan client)
)

func broadcast() {
	clients := make(map[client]bool)
	var playerOrder []client
	playerTurn := 0
	for {
		select {
		case cli := <-entering:
			clients[cli] = true
			if len(playerOrder) == 0 {
				m, err := json.Marshal(Message{Start: true})
				if err != nil {
					// TODO
					log.Fatal(err)
				}
				cli <- string(m)
			}
			playerOrder = append(playerOrder, cli)
		case msg := <-messages:
			if len(playerOrder) == 0 {
				continue
			}
			log.Print("PLAYERS!!! <3 ", playerTurn, playerOrder)
			c := playerOrder[playerTurn]
			// fmt.Printf("MESSAGE <3: %s\n", msg)
			for cli := range clients {
				if c == cli {
					cli <- msg
					if playerTurn == len(playerOrder)-1 {
						playerTurn = 0
					} else {
						playerTurn++
					}
				} else {
					e, err := json.Marshal(Message{EmptyLine: true})
					if err != nil {
						// TODO
						log.Fatal(err)
					}
					cli <- string(e)
				}
			}
		case cli := <-leaving:
			delete(clients, cli)
			for i, c := range playerOrder {
				if c == cli {
					playerOrder = append(playerOrder[:i], playerOrder[i+1:]...)
					if playerTurn == 0 {
						playerTurn = len(playerOrder) - 1
					} else {
						playerTurn--
					}
				}
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

	ch := make(chan string)
	go clientWriter(ws, ch)

	who := ws.conn.RemoteAddr().String()
	j, err := json.Marshal(Message{Join: who})
	if err != nil {
		// TODO
		log.Fatal(err)
	}
	e, err := json.Marshal(Message{Entering: who})
	log.Print("HELLO ENTERING <3:")
	log.Print(string(e))
	if err != nil {
		// TODO
		log.Fatal(err)
	}

	ch <- string(j)
	messages <- string(e)
	entering <- ch

loop:
	for {
		msg, opcode, err := ws.Recv()

		if err != nil {
			switch err {
			case io.EOF:
				fmt.Println("end of file <3")
				break loop
			default:
				fmt.Println("closing error <3")
				fmt.Println(err.Error())
				break loop
			}
		} else if opcode == 0x9 {
			// If receive PING, send PONG back
			// if the connection wasn't closed,
			// TODO: sending back the same Application
			// Data from the PING
			ws.Pong()
		}

		// Make sure to broadcast only text messages,
		// not Control frames like Close, Ping, and Pong
		if opcode == 1 {
			log.Println(msg)
			messages <- string(msg)
		}
	}

	l, err := json.Marshal(Message{Leaving: who})
	if err != nil {
		// TODO
		log.Fatal(err)
	}
	leaving <- ch
	messages <- string(l)
	return nil
}

func clientWriter(ws *Ws, ch <-chan string) {
	for msg := range ch {
		err := ws.SendMsg(msg)
		if err != nil {
			log.Fatal(err)
			// Send HTTP error code
			ws.conn.Close()
		}
	}
}

func handshake(w http.ResponseWriter, r *http.Request) (*Ws, error) {
	httpStatus, err := validateWsRequest(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), httpStatus)
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
	fmt.Printf("host: %s\n", host)
	if host != "localhost:8080" {
		return http.StatusForbidden, fmt.Errorf("forbidden host: %s\n", host)
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
		return http.StatusInternalServerError, errors.New("error decoding Sec-WebSocket-Key header")
	}
	if len(wsKeyBytes) != 16 {
		return http.StatusBadRequest, fmt.Errorf("invalid Sec-WebSocket-Key header length of %d, must be 16-bytes long", len(wsKeyBytes))
	}

	return 0, nil
}

func (ws *Ws) Close() {
	err := ws.Send("", 0x8)
	if err != nil {
		log.Fatal(err)
	}

	err = ws.conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (ws *Ws) SendMsg(msg string) error {
	err := ws.Send(msg, 1)
	return err
}

func (ws *Ws) write(data []byte) error {
	_, err := ws.bufrw.Write(data)
	if err != nil {
		return err
	}
	return ws.bufrw.Flush()
}

func (ws *Ws) read(buf []byte) error {
	_, err := ws.bufrw.Read(buf)
	if err != nil {
		return err
	}
	return nil
}

func (ws *Ws) Recv() (string, uint8, error) {
	// TODO: opcode, fail if RSV values are not 0,
	// fail if not masked, unmask
	head1 := make([]byte, 2)
	err := ws.read(head1)
	if err != nil {
		return "", 0, err
	}
	parsedFrame := parseFrameHead(head1)
	// TODO: validate rsv1, rsv2, rsv3, and mask
	if parsedFrame.payLen == 126 {
		head2 := make([]byte, 2)
		err = ws.read(head2)
		if err != nil {
			return "", 0, err
		}
		var byte1 = uint16(head2[0])
		var byte2 = uint16(head2[1])
		parsedFrame.extPayLen1 = (byte1 << 8) | byte2
	}
	if parsedFrame.payLen == 127 {
		head2 := make([]byte, 8)
		err = ws.read(head2)
		if err != nil {
			return "", 0, err
		}
		var acc uint64
		for i := 0; i < 8; i++ {
			shift := 8 * (7 - i)
			next := uint64(head2[i])
			acc = (next << shift) | acc
		}
		parsedFrame.extPayLen2 = acc
	}
	maskKey := make([]byte, 4)
	err = ws.read(maskKey)
	if err != nil {
		return "", 0, err
	}

	parsedFrame.maskKey = maskKey

	payLen := getPayloadLength(parsedFrame)
	pay := make([]byte, payLen)

	err = ws.read(pay)
	if err != nil {
		return "", 0, err
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
	unmasked := make([]byte, frame.payLen)
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
	log.Println(byte1)
	err := frame.WriteByte(byte1)
	if err != nil {
		return err
	}

	switch {
	case payLen < 126:
		// 7 bits to denote payload length (in bytes)
		byte2 := (masked << 7) | uint8(payLen)
		err = frame.WriteByte(byte2)
		if err != nil {
			return err
		}
	case payLen < (1 << 16):
		// 7+16 bits
		byte2 := (masked << 7) | (uint8(126))
		err = frame.WriteByte(byte2)
		if err != nil {
			return err
		}
		bytes34 := uint16(payLen)
		err = binary.Write(frame, binary.BigEndian, bytes34)
		if err != nil {
			return err
		}
	default:
		// 7+64 bits
		byte2 := (masked << 7) | (uint8(127))
		err = frame.WriteByte(byte2)
		if err != nil {
			return err
		}
		nextBytes := uint64(payLen)
		err = binary.Write(frame, binary.BigEndian, nextBytes)
		if err != nil {
			return err
		}
	}
	_, err = frame.Write(pay)
	if err != nil {
		return err
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
