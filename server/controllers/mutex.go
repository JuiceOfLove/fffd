package controllers

import (
	"github.com/gofiber/websocket/v2"
	"sync"
)

// TicketRooms хранит все WebSocket-соединения для каждого тикета: ticketID → (conn → userID)
var TicketRooms   = make(map[uint]map[*websocket.Conn]uint)

// mutex для безопасного доступа к TicketRooms
var TicketRoomsMu sync.Mutex
