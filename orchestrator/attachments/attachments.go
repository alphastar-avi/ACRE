package attachments

// Attachment represents an image or other file attached to the ticket.
type Attachment struct {
	Path string
}

// Load retrieves all attachments associated with a ticket.
// For the MVP, this just returns an empty list as we are ignoring images.
func Load(ticketDir string) ([]Attachment, error) {
	// Stub implementation for MVP
	return []Attachment{}, nil
}
