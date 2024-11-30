package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mekdep/server/config"
)

type ShortMessage struct {
	Phones []string `json:"phones"`
	Text   string   `json:"text"`
}

func SendMessageToPhones(phones []string, text string) error {
	smppServerURL := config.Conf.SmppServerURL + "/api/v0/messages"
	data := &ShortMessage{
		Phones: phones,
		Text:   text,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", smppServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", config.Conf.SmppServerToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}
