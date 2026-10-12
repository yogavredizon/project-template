package rabbitmq

type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

type Message struct {
	Exchange   string
	RoutingKey string
	Headers    map[string]any
	Priority   uint8
	Body       Encoder
}

type Delivery struct {
	body        []byte
	contentType string
}

func (d Delivery) Encode() ([]byte, string, error) {
	return d.body, d.contentType, nil
}
