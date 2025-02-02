package notify

// Client -
type Client struct {
	C chan []byte
}

func (c *Client) Write(payload []byte) (int, error) {
	c.C <- payload

	return len(payload), nil
}

// New -
func New() *Client {
	return &Client{
		C: make(chan []byte, 20),
	}
}
