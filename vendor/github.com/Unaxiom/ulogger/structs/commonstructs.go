package commonstructs

// LogMessage is the internal struct that is posted to the remote log server
type LogMessage struct {
	MessageType      string `json:"message_type"`
	Timestamp        int64  `json:"timestamp"`
	MessageContent   string `json:"message_content"`
	OrganizationName string `json:"organization_name"`
	ApplicationName  string `json:"application_name"`
	File             string `json:"file"`
	Function         string `json:"function"`
	Line             int64  `json:"line"`
}

// PostMessage is the struct that will be used to send a message to the remote URL
type PostMessage struct {
	MessageTag string       `json:"message_tag"`
	LogList    []LogMessage `json:"log_list"`
}
