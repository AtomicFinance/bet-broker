// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package broker

import (
	"log"
)

type Message struct {
	From string
	To   string
	Body string
}

type IncomingMessage struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

type OutgoingMessage struct {
	From string `json:"from"`
	Body string `json:"body"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
			}
		case wrappedMessage := <-h.broadcast:
			if client, ok := h.clients[wrappedMessage.To]; ok {
				select {
				case client.send <- OutgoingMessage{wrappedMessage.From, wrappedMessage.Body}:
				default:
					close(client.send)
					delete(h.clients, wrappedMessage.To)
				}
			} else {
				log.Println("err: client not found")
			}
		}
	}
}
