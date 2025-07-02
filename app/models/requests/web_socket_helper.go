package requests

import (
	"encoding/json"
	"users-service/app/models/responses"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type WebSocketHelper struct {
	Conn   Conn
	Logger zerolog.Logger
}

type Conn interface {
	WriteJSON(v interface{}) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

type WebSocketHelperInterface interface {
	ForwardProgress(progressChan <-chan responses.Message)
	SendMessage(msg responses.Message)
	RespondFailure(message, errorDetail string)
	RespondSuccess(message, requestID string, data interface{})
	ReadRequest(v interface{}) error
}

func (h *WebSocketHelper) ForwardProgress(progressChan <-chan responses.Message) {
	go func() {
		for msg := range progressChan {
			h.SendMessage(msg)
		}
	}()
}

func (h *WebSocketHelper) SendMessage(msg responses.Message) {
	if err := h.Conn.WriteJSON(msg); err != nil {
		h.Logger.Error().Err(err).Msg("Error sending WebSocket message")
	}
}

func (ws *WebSocketHelper) RespondFailure(message, errorDetail string) {
	ws.SendMessage(responses.Message{
		Status:  "error",
		Message: message,
		Error:   errorDetail,
	})
}

func (ws *WebSocketHelper) RespondSuccess(message, requestID string, data interface{}) {
	ws.SendMessage(responses.Message{
		Status:    "success",
		Message:   message,
		Data:      data,
		RequestID: requestID,
	})
}

func (h *WebSocketHelper) ReadRequest(v interface{}) error {
	messageType, p, err := h.Conn.ReadMessage()
	if err != nil {
		h.Logger.Error().Err(err).Msg("Error reading message from WebSocket")
		return err
	}

	if messageType != websocket.TextMessage {
		h.SendMessage(responses.Message{Status: "error", Message: "Unsupported message type"})
		h.Logger.Warn().Msg("Received unsupported WebSocket message type")
		return err
	}

	if err := json.Unmarshal(p, v); err != nil {
		h.SendMessage(responses.Message{Status: "error", Message: "Invalid data format", Error: err.Error()})
		h.Logger.Error().Err(err).Msg("Failed to unmarshal data from WebSocket")
		return err
	}

	return nil
}
