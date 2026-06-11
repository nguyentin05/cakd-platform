package notify

// Notifier defines the interface for provisioning notification channels and broadcasting alerts.
// Implementations of this interface manage chat integration channels (such as Discord).
type Notifier interface {
	// ProvisionChannel creates a dedicated workspace channel and returns its webhook URL endpoint.
	ProvisionChannel(projectName string) (webhookURL string, err error)
	// SendAlert sends the structured alert payload to the specified webhook URL.
	SendAlert(webhookURL string, payload AlertPayload) error
}

// AlertPayload holds a batch of individual alert items to be dispatched to notification channels.
type AlertPayload struct {
	// Items represents the list of alert details.
	Items []AlertItem
}

// AlertItem represents a single system diagnostic alert containing status descriptions.
type AlertItem struct {
	Title       string
	Description string
	Severity    string
}
