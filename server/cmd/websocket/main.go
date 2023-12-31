package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	http.HandleFunc("/ws", handleWebsocket)
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func handleWebsocket(w http.ResponseWriter, req *http.Request) {
	u := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	c, err := u.Upgrade(w, req, nil)
	if err != nil {
		fmt.Printf("cannot upgrade: %+v", err)
		return
	}
	defer c.Close()

	done := make(chan struct{})
	// 接消息
	go func() {
		for {
			m := make(map[string]interface{})
			err := c.ReadJSON(&m)
			if err != nil {
				// 处理断开连接
				if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					fmt.Printf("unexpected read error: %v\n", err)
				}
				done <- struct{}{}
				break
			}
			fmt.Printf("message received: %v\n", m)
		}
	}()
	// 发消息
	i := 0
	for {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-done:
			return
		}

		i++
		err := c.WriteJSON(map[string]string{
			"hello":  "websocket",
			"msg_id": strconv.Itoa(i),
		})
		if err != nil {
			fmt.Printf("cannot write json: %v\n", err)
		}
	}
}
