package tracking

// Event represents an event for tracking
type Event struct {
	ID         string
	Name       string
	Properties map[string]interface{}
}
