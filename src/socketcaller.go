package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ip          = flag.String("ip", "127.0.0.1", "server IP")
	connections = flag.Int("conn", 1000, "number of websocket connections")
)

// Payload struct save payload from client
type Payload struct {
	Domain   string `json:"domain"`
	Session  string `json:"session"`
	UpdateAt string `json:"updateat"`
}

func main() {
	flag.Usage = func() {
		io.WriteString(os.Stderr, `Websockets client generator Example usage: ./client -ip=172.17.0.1 -conn=10`)
		flag.PrintDefaults()
	}
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *ip + ":8880", Path: "/ping"}
	log.Printf("Connecting to %s", u.String())
	var conns []*websocket.Conn
	for i := 0; i < *connections; i++ {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			fmt.Println("Failed to connect", i, err)
			break
		}
		conns = append(conns, c)
		defer func() {
			c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
			time.Sleep(time.Second)
			c.Close()
		}()
	}

	log.Printf("Finished initializing %d connections", len(conns))
	tts := time.Second
	if *connections > 100 {
		tts = time.Second * 30
	}
	for {
		for i := 0; i < len(conns); i++ {
			conn := conns[i]
			log.Printf("Conn %d sending message", i)
			location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
			now := time.Now().In(location).Format("2006-01-02 15:04:05")
			var payload = Payload{Domain: "localhost", Session: fmt.Sprintf("%v", i), UpdateAt: now}

			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(payload)

			conn.WriteMessage(websocket.TextMessage, reqBodyBytes.Bytes())
		}
		log.Printf("Finish send from %d socket", len(conns))
		time.Sleep(tts)
	}
}
