package campaign

import (
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"io"
	"log"
)

var stream = NewServer()

type Message struct {
	CampaignID              CampaignID              `json:"campaignID"`
	CampaignEtablissementID CampaignEtablissementID `json:"campaignEtablissementID"`
	Zone                    []string                `json:"zone"`
	Type                    string                  `json:"type"`
	Username                string                  `json:"username"`
}

func (m Message) JSON() string {
	msg, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(msg)
}

func streamHandler(c *gin.Context) {
	var s core.Session
	s.Bind(c)

	v, ok := c.Get("clientChan")
	if !ok {
		return
	}
	clientChan, ok := v.(ClientChan)
	if !ok {
		return
	}

	zone := zoneForUser(libwekan.Username(s.Username))

	c.Stream(func(w io.Writer) bool {
		// Stream message to client from message channel
		if message, ok := <-clientChan; ok {
			if utils.Overlaps(zone, message.Zone) {
				c.SSEvent("message", message.JSON())
			}
			return true
		}
		return false
	})
}

// It keeps a list of clients those are currently attached
// and broadcasting events to those clients.
type Event struct {
	// Events are pushed to this channel by the main events-gathering routine
	Message chan Message

	// New client connections
	NewClients chan chan Message

	// Closed client connections
	ClosedClients chan chan Message

	// Total client connections
	TotalClients map[chan Message]bool
}

// New event messages are broadcast to all registered client connection channels
type ClientChan chan Message

// Initialize event and Start procnteessing requests
func NewServer() (event *Event) {
	event = &Event{
		Message:       make(chan Message),
		NewClients:    make(chan chan Message),
		ClosedClients: make(chan chan Message),
		TotalClients:  make(map[chan Message]bool),
	}

	go event.listen()

	return
}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (stream *Event) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Message:
			for clientMessageChan := range stream.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

func (stream *Event) serveHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		stream.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- clientChan
		}()

		c.Set("clientChan", clientChan)

		c.Next()
	}
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}