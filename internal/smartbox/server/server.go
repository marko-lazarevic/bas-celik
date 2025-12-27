// Package server implements the Smartbox server functionality.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/ubavic/bas-celik/v2/internal/logger"
)

// Message represents a generic message structure for Smartbox communication.
type Message[I any] struct {
	Operation string `json:"operation"`
	Input     I
}

// Response represents a generic response structure for Smartbox communication.
type Response[P any] struct {
	Operation string `json:"operation"`
	Status    int    `json:"status"`
	Payload   P      `json:"payload"`
}

// OnOpenPayload represents the payload sent on connection open.
type OnOpenPayload struct {
	AppName    string `json:"appName"`
	AppVersion string `json:"appVersion"`
	AppBuild   int    `json:"appBuild"`
	OsName     string `json:"osName"`
	OsVersion  string `json:"osVersion"`
	OsArch     string `json:"osArch"`
}

var onOpenEvent = "ON_OPEN_EVENT"
var operationGetInfo = "GET_INFO"
var operationGetProviders = "GET_PROVIDERS"
var operationGetTerminals = "GET_TERMINALS"
var operationGetCertificates = "GET_CERTIFICATES"
var operationGetSignedXML = "GET_SIGNED_XML"

// SmartBoxServer represents the Smartbox server.
type SmartBoxServer struct {
	sessions    map[string]SmartboxSession
	modulePaths []ModulePath
}

// ServeHTTP handles incoming HTTP requests and upgrades them to WebSocket connections.
func (s *SmartBoxServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"eporezi.purs.gov.rs"},
	})

	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.CloseNow()

	err = s.smartBoxHandler(conn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(err)
	}
}

func (s *SmartBoxServer) smartBoxHandler(conn *websocket.Conn) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelCtx()

	w, err := conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		return err
	}

	err = onOpen(w)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	sessionID := ""

	for {
		err = s.onMessage(&sessionID, ctx, conn)
		if err != nil {
			logger.Error(err)
			break
		}
	}

	return conn.CloseNow()
}

func (s *SmartBoxServer) onMessage(sessionID *string, ctx context.Context, conn *websocket.Conn) error {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return err
	}

	msg := Message[any]{}
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("Received `%s` message. Session `%s`.", msg.Operation, *sessionID))

	w, err := conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		return err
	}

	defer w.Close()

	session, ok := s.sessions[*sessionID]
	if !ok && msg.Operation != operationGetInfo {
		return fmt.Errorf("session not found")
	}

	switch msg.Operation {
	case operationGetInfo:
		err = s.handleGetInfo(sessionID, data, w)
	case operationGetProviders:
		err = s.handleGetProviders(data, w)
	case operationGetTerminals:
		err = s.handleGetTerminals(&session, data, w)
	case operationGetCertificates:
		err = s.handleGetCertificates(&session, data, w)
	case operationGetSignedXML:
		err = s.handleGetSignedXML(&session, data, w)
	default:
		err = fmt.Errorf("unknown operation %s", msg.Operation)
	}

	s.sessions[*sessionID] = session
	return err
}
