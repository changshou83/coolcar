package ws

import (
	"context"
	"coolcar/car/mq"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func CreateHandler(u *websocket.Upgrader, sub mq.Subscriber, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		c, err := u.Upgrade(w, req, nil)
		if err != nil {
			logger.Warn("cannot upgrade", zap.Error(err))
			return
		}
		defer c.Close()

		msgs, cleanUp, err := sub.Subscribe(context.Background())
		defer cleanUp()
		if err != nil {
			logger.Error("cannot subscribe", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		done := make(chan struct{})
		// receive messages
		go func() {
			for {
				_, _, err := c.ReadMessage()
				if err != nil {
					// if not close
					if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						logger.Error("unexpected read error", zap.Error(err))
					}
					done <- struct{}{}
					break
				}
			}
		}()

		// send messages
		for {
			select {
			case msg := <-msgs:
				err := c.WriteJSON(msg)
				if err != nil {
					logger.Warn("cannot write JSON", zap.Error(err))
				}
			case <-done:
				return
			}
		}
	}
}
