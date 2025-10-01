package uploader

import (
	"bytes"
	"chunks/utils"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	cfg := utils.ReadCfg()
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var chunksNum int
	if header.Size%int64(cfg.ChunkSize) != 0 {
		chunksNum = int(math.Floor(float64(header.Size/int64(cfg.ChunkSize)))) + 1
	} else {
		chunksNum = int(math.Floor(float64(header.Size / int64(cfg.ChunkSize))))
	}

	links := make([]string, chunksNum+2)

	links[0] = header.Filename
	links[1] = fmt.Sprintf("%d", header.Size)

	s := utils.NewSemaphore(uint(cfg.MaxRoutine))
	for i := range chunksNum {
		buffer := make([]byte, cfg.ChunkSize)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		fmt.Println("Byte read: ", n)

		if n == 0 {
			fmt.Println("read all file done")
			break
		}

		s.Add()
		go func(index int) {
			defer func() {
				s.Done()
				fmt.Printf("finished %d\n", index)
			}()

			link, err := uploadChunkToDiscord(utils.XOR1(buffer[:n], cfg.Key), header.Filename, cfg)
			if err != nil {
				http.Error(w, "Error uploading chunk to Discord: "+err.Error(), http.StatusInternalServerError)
				return
			}

			links[index+2] = strings.ReplaceAll(link, "https://cdn.discordapp.com/attachments/", "")
		}(i)
	}

	s.Wait()

	fileInfoLink := strings.ReplaceAll(
		strings.ReplaceAll(writeOutputLinks(links, cfg), "https://cdn.discordapp.com/attachments/", ""),
		"/LOD.LOD",
		"",
	)

	fmt.Fprintf(w, "/download/%s", fileInfoLink)
}

func uploadChunkToDiscord(data []byte, originalFileName string, cfg utils.Config) (string, error) {
	webhooks := cfg.GetWebhooks()
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	fileWriter, err := writer.CreateFormFile("file", originalFileName)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(data)
	_, err = io.Copy(fileWriter, reader)
	if err != nil {
		return "", err
	}
	writer.Close()

	resp, err := http.Post(utils.GetRandomWebhook(webhooks), writer.FormDataContentType(), &requestBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	//fmt.Printf("Response: %s\n", responseBody)
	//fmt.Printf("Response status: %s\n", resp.Status)

	var a utils.DiscordWebhook
	if err := json.Unmarshal(responseBody, &a); err != nil {
		return "", err
	}

	fmt.Println(string(responseBody))

	attachmentUrl, _ := url.Parse(a.Attachments[0].URL)

	fmt.Println(a.Attachments[0].URL)
	return fmt.Sprintf("https://%s%s", attachmentUrl.Host, attachmentUrl.Path), nil
}

func writeOutputLinks(links []string, cfg utils.Config) string {
	var output bytes.Buffer
	for _, link := range links {
		output.WriteString(link + "\n")
	}

	link, _ := uploadChunkToDiscord(output.Bytes(), "LOD.LOD", cfg)
	return link
}
