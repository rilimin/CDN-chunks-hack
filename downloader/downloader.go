package downloader

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"chunks/utils"
)

type chunk struct {
	index uint
	data  bytes.Buffer
}

func download(link string) bytes.Buffer {
	var ret bytes.Buffer
	resp, err := http.Get(link)
	utils.Check(err)
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
	}

	_, err = io.Copy(&ret, resp.Body)
	utils.Check(err)
	return ret
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	cfg := utils.ReadCfg()
	getFileInfo := func() utils.FileInfo {
		//var fileInfo FileInfo
		var info bytes.Buffer

		resp, err := http.Get(
			utils.RenewLinks([]string{
				fmt.Sprintf("https://cdn.discordapp.com/attachments/%s/%s/LOD.LOD",
					r.PathValue("id1"),
					r.PathValue("id2"),
				)})[0])

		utils.Check(err)
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			fmt.Println(resp.StatusCode)
		}

		_, err = io.Copy(&info, resp.Body)
		utils.Check(err)
		return utils.CreateFileInfo(info.String())
	}

	fileInfo := getFileInfo()
	newLinks := utils.RenewLinks(fileInfo.Links)

	totalChunks := len(newLinks)
	currentChunk := 0

	w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.FileName)
	w.Header().Set("Content-Length", fileInfo.Size)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Expires", "0")

	s := utils.NewSemaphore(cfg.MaxRoutine)

	fmt.Println("Total chunk: ", totalChunks)
	for {
		if currentChunk < totalChunks {
			chunkChan := make(chan chunk, cfg.MaxRoutine)

			for i := range cfg.MaxRoutine {
				if currentChunk >= totalChunks {
					break
				}

				s.Add()
				go func(index int) {
					defer func() {
						s.Done()
						fmt.Printf("Done %d/%d\n", index, totalChunks)
					}()

					data := download(newLinks[index])
					chunkChan <- chunk{index: i, data: data}
				}(currentChunk)
				currentChunk++
			}

			go func() {
				s.Wait()
				close(chunkChan)
			}()

			chunks := make([]bytes.Buffer, cfg.MaxRoutine)
			var written uint = 0

			for c := range chunkChan {
				chunks[c.index] = c.data

				for written < cfg.MaxRoutine && chunks[written].Bytes() != nil {
					if _, err := io.Copy(w, utils.XOR2(&chunks[written], byte(0x69))); err != nil {
						return
					}
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
					written++
				}
			}

			fmt.Println("written ", written)
		} else {
			break
		}
	}
}
