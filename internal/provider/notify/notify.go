package notify

// Notifier manages notification channels and sends alerts.
type Notifier interface {
	ProvisionChannel(projectName string) (webhookURL string, err error)
	SendAlert(webhookURL string, payload AlertPayload) error
}

// AlertPayload contains a list of alert items to send.
type AlertPayload struct {
	Items []AlertItem
}

// AlertItem represents a single alert notification.
type AlertItem struct {
	Title       string
	Description string
	Severity    string
}
