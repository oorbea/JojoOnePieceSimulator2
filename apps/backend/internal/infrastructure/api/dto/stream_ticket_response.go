package dto

import "time"

// StreamTicketResponse is the JSON body returned by both stream-ticket mint
// routes (POST /events/ticket and POST /games/{id}/ws-ticket): a short-lived,
// single-use Ticket the caller appends as ?ticket=<...> to the actual
// EventSource/WebSocket URL.
type StreamTicketResponse struct {
	Ticket    string    `json:"ticket"`
	ExpiresAt time.Time `json:"expiresAt"`
}
