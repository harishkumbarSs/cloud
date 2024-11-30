package ws

import (
    "encoding/json"
    "log"
    "sync"

    "github.com/gorilla/websocket"
)

type Client struct {
    UserEmail string
    Conn      *websocket.Conn
}

type Message struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

var (
    clients    = make(map[*Client]bool)
    clientsMux sync.RWMutex
)

func AddClient(client *Client) {
    clientsMux.Lock()
    clients[client] = true
    clientsMux.Unlock()
}

func RemoveClient(client *Client) {
    clientsMux.Lock()
    delete(clients, client)
    clientsMux.Unlock()
}

func BroadcastToUser(userEmail string, message Message) {
    clientsMux.RLock()
    defer clientsMux.RUnlock()

    messageJSON, err := json.Marshal(message)
    if err != nil {
        log.Printf("Error marshaling message: %v", err)
        return
    }

    for client := range clients {
        if client.UserEmail == userEmail {
            err := client.Conn.WriteMessage(websocket.TextMessage, messageJSON)
            if err != nil {
                log.Printf("Error sending message to client: %v", err)
                client.Conn.Close()
                RemoveClient(client)
            }
        }
    }
}
