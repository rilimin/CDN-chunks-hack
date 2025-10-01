package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
)

func Check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func XOR1(data []byte, key byte) []byte {
	result := make([]byte, len(data))
	for i := range len(data) {
		result[i] = data[i] ^ key
	}
	return result
}

func XOR2(data *bytes.Buffer, key byte) *bytes.Buffer {
	result := bytes.NewBuffer(nil) // Create a new buffer for the result.

	// Process each byte in the input buffer.
	for {
		b, err := data.ReadByte()
		if err != nil {
			break // Exit the loop when we reach the end of the buffer.
		}
		result.WriteByte(b ^ key) // XOR the byte and write it to the result buffer.
	}

	return result
}

func CreateFileInfo(info string) FileInfo {
	lines := strings.Split(info, "\n")

	var links []string
	for i := 0; i < len(lines)-3; i++ {
		links = append(links, fmt.Sprintf("https://cdn.discordapp.com/attachments/%s", lines[i+2]))
	}

	return FileInfo{
		FileName: lines[0],
		Size:     lines[1],
		Links:    links,
	}
}

func RenewLinks(links []string) []string {
	cfg := ReadCfg()
	discordToken := cfg.UserToken
	var ret []string
	if len(links) < 40 {
		requestBody := RefreshRequest{
			AttachmentURLs: links,
		}

		jsonBody, err := json.Marshal(requestBody)
		Check(err)

		req, err := http.NewRequest("POST", "https://discord.com/api/v9/attachments/refresh-urls", bytes.NewBuffer(jsonBody))
		Check(err)

		req.Header.Set("Authorization", discordToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		Check(err)

		defer resp.Body.Close()

		var response RefreshResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Println(err)
		}

		for _, refreshed := range response.RefreshedUrls {
			ret = append(ret, refreshed.Refreshed)
		}
	} else {
		for i := 0; i < len(links); i += 40 {
			end := i + 40
			if end > len(links) {
				end = len(links)
			}
			batch := links[i:end]

			requestBody := RefreshRequest{
				AttachmentURLs: batch,
			}

			jsonBody, err := json.Marshal(requestBody)
			Check(err)

			req, err := http.NewRequest("POST", "https://discord.com/api/v9/attachments/refresh-urls", bytes.NewBuffer(jsonBody))
			Check(err)

			req.Header.Set("Authorization", discordToken)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			Check(err)

			defer resp.Body.Close()

			var response RefreshResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				fmt.Println(err)
			}

			for _, refreshed := range response.RefreshedUrls {
				ret = append(ret, refreshed.Refreshed)
			}
		}
	}

	fmt.Println(ret)
	return ret
}

func GetRandomWebhook(webhooks []string) string {
	return webhooks[rand.IntN(len(webhooks))]
}
