package signal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	Base     string // e.g., http://host:8080
	Room     string // meeting ID
	ClientID string // uuid or random string
	HC       *http.Client
}

func New(base, room, clientID string) *Client {
	return &Client{Base: base, Room: room, ClientID: clientID, HC: &http.Client{}}
}

func (c *Client) Join() (string, error) {
	body, _ := json.Marshal(map[string]string{"clientId": c.ClientID})
	resp, err := c.HC.Post(c.Base+"/room/"+c.Room+"/join", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("join status %s", resp.Status)
	}
	var v struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}
	return v.Role, nil
}

func (c *Client) PostOffer(sdp string) error {
	body, _ := json.Marshal(map[string]string{"sdp": sdp})
	resp, err := c.HC.Post(c.Base+"/room/"+c.Room+"/offer?clientId="+c.ClientID, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("post offer failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) GetOffer() (string, bool, error) {
	resp, err := c.HC.Get(c.Base + "/room/" + c.Room + "/offer")
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", false, nil
	}
	var v struct {
		SDP string `json:"sdp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", false, err
	}
	return v.SDP, true, nil
}

func (c *Client) PostAnswer(sdp string) error {
	body, _ := json.Marshal(map[string]string{"sdp": sdp})
	resp, err := c.HC.Post(c.Base+"/room/"+c.Room+"/answer?clientId="+c.ClientID, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("post answer failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) GetAnswer() (string, bool, error) {
	resp, err := c.HC.Get(c.Base + "/room/" + c.Room + "/answer")
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", false, nil
	}
	var v struct {
		SDP string `json:"sdp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", false, err
	}
	return v.SDP, true, nil
}

func (c *Client) PostICE(from, candB64 string) error {
	body, _ := json.Marshal(map[string]string{"candidate": candB64})
	resp, err := c.HC.Post(c.Base+"/room/"+c.Room+"/ice?from="+from+"&clientId="+c.ClientID, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("post ice failed: %s", resp.Status)
	}
	return nil
}

// Subscribe ICE (SSE) to="offer" or "answer"
func (c *Client) SubscribeICE(to string, onCand func(string)) (*http.Response, error) {
	resp, err := c.HC.Get(c.Base + "/room/" + c.Room + "/ice?to=" + to)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("sse status %s", resp.Status)
	}
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			line := sc.Text()
			if len(line) > 6 && line[:6] == "data: " {
				onCand(line[6:])
			}
		}
		resp.Body.Close()
	}()
	return resp, nil
}
