package cabbage

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// CabbageMessage base message for publish\consume
type CabbageMessage struct {
	ID        string    `json:"id"`
	MessageId string    `json:"messageId"`
	Body      []byte    `json:"body"`
	TaskName  string    `json:"TaskName"`
	Timestamp time.Time `json:"timestamp"`
}

// newCabbageMessage create cabbage message
func newCabbageMessage(taskName string, body []byte) *CabbageMessage {
	cbMessage := &CabbageMessage{
		ID:        uuid.NewV4().String(),
		MessageId: uuid.NewV4().String(),
		Body:      body,
		TaskName:  taskName,
		Timestamp: time.Now(),
	}
	return cbMessage
}
