package ticket

import (
	"encoding/json"
	"fmt"
	"os"
)

type Comment struct {
	Created string `json:"created"`
	Author  string `json:"author"`
	Body    string `json:"body"`
}

// Ticket represents the structure of an incident ticket JSON.
type Ticket struct {
	TicketID           string    `json:"ticket_id"`
	Summary            string    `json:"summary"`
	Description        string    `json:"description"`
	AcceptanceCriteria string    `json:"acceptance_criteria"`
	Comments           []Comment `json:"comments"`
}

// Load reads and parses a ticket JSON file.
func Load(filePath string) (*Ticket, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ticket file: %w", err)
	}

	var t Ticket
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("failed to parse ticket JSON: %w", err)
	}

	return &t, nil
}
