package downloader

import (
	"bytes"
	"chunks/utils"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

const (
	sizePerRequst = 10 * 1024 * 1023 // 10MB/req
)

var fileInfo utils.FileInfo
var newLinks []string
var size int64
var id1, id2 string

func RangeVideo(w http.ResponseWriter, req *http.Request) {
	//f, size, err := openfile(vname)

	fmt.Println("--- Request Headers ---")
	for name, values := range req.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}
	fmt.Println("Method: " + req.Method)
	fmt.Println("-----------------------")

	getFileInfo := func() (utils.FileInfo, error) {
		//var fileInfo FileInfo
		var info bytes.Buffer

		resp, err := http.Get(
			utils.RenewLinks([]string{
				fmt.Sprintf("https://cdn.discordapp.com/attachments/%s/%s/LOD.LOD",
					id1,
					id2,
				)})[0])

		if err != nil {
			return utils.FileInfo{}, fmt.Errorf("failed to send get request, error: %s", err.Error())
		}
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return utils.FileInfo{}, fmt.Errorf("failed to request file info, response: %d", resp.StatusCode)
		}

		_, err = io.Copy(&info, resp.Body)
		if err != nil {
			return utils.FileInfo{}, fmt.Errorf("failed to copy response body, error: %s", err.Error())
		}
		return utils.CreateFileInfo(info.String()), nil
	}

	//size, _ := strconv.ParseInt(fileInfo.Size, 10, 64)

	if id1 != req.PathValue("id1") || id2 != req.PathValue("id2") {
		id1 = req.PathValue("id1")
		id2 = req.PathValue("id2")

		fileInfo, err := getFileInfo()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newLinks = utils.RenewLinks(fileInfo.Links)
		size, _ = strconv.ParseInt(fileInfo.Size, 10, 64)
	}

	w.Header().Set("Content-Type", "video/mp4")

	rangeHeader := req.Header.Get("Range")
	// we can simply hint Chrome to send serial range requests for media file by
	//
	// if rangeHeader == "" {
	// 	w.Header().Set("Accept-Ranges", "bytes")
	// 	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	// 	w.WriteHeader(200)
	// 	fmt.Printf("hint browser to send range requests, total size: %d\n", size)
	// 	return
	// }
	//
	// but this not worked for Safari and Firefox
	if rangeHeader == "" {

		ra := httpRange{
			start:  0,
			length: sizePerRequst,
		}
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", strconv.FormatInt(ra.length, 10))
		w.Header().Set("Content-Range", fileInfo.Size)
		w.WriteHeader(http.StatusPartialContent)
		fmt.Printf("hint browser to send serial range requests, response 206, 0-%d/%d\n", sizePerRequst-1, size)
		if req.Method != "HEAD" {
			// write first 10MB of video file
			data := download(newLinks[0])
			_, err := io.Copy(w, utils.XOR2(&data, byte(0x69)))

			//if written != ra.length {
			//	fmt.Printf("desired range size: %d, actual written: %d, err: %v\n\n", ra.length, written, err)
			//}

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error writing response: %v\n", err)
			} else {
				fmt.Println("response written successfully")
			}
		}
		return
	}

	// browser sends range request
	reqer := req.RemoteAddr
	fmt.Printf("\n%s request range %s\n", reqer, rangeHeader)
	ranges, err := parseRange(rangeHeader, size)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	// multi-part requests are not supported
	if len(ranges) > 1 {
		http.Error(w, "unsuported multi-part", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	ra := ranges[0]

	var chunksNum int
	tempStart2 := ra.start + 1
	//chunkSize := 10 * 1024 * 1023

	if tempStart2%int64(sizePerRequst) != 0 {
		chunksNum = int(tempStart2 / int64(sizePerRequst))
	} else if ra.start == 0 {
		chunksNum = 0
	} else {
		chunksNum = int(tempStart2/int64(sizePerRequst)) - 1
	}

	/*
		if _, err := f.Seek(ra.start, io.SeekStart); err != nil {
			http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
			return
		}
	*/

	tempStart := max(int(ra.start)-chunksNum*sizePerRequst, 0)

	sendSize := ra.length - int64(tempStart)
	if sendSize < 0 {
		sendSize = ra.length
	}
	fmt.Printf("response range bytes %d-%d, %d KB\n", ra.start, ra.start+ra.length-1, ra.length/1024)
	fmt.Printf("chunksNum: %d, tempStart: %d SendSize: %d\n", chunksNum, tempStart, sendSize)

	w.Header().Set("Content-Range", ra.contentRange(size))
	w.Header().Set("Accept-Ranges", "bytes")
	if w.Header().Get("Content-Encoding") == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
	}
	w.WriteHeader(http.StatusPartialContent)

	if req.Method != "HEAD" {
		data := download(newLinks[chunksNum])
		decryptedData := utils.XOR2(&data, byte(0x69))

		written, err := io.CopyN(w, bytes.NewBuffer(decryptedData.Bytes()[tempStart:]), sendSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if written != sendSize || err != nil {
			fmt.Printf("desired range size: %d, actual written: %d, err: %v\n\n", sendSize, written, err)
		} else {
			fmt.Println()
		}
	}
}

// --- httpRange and its funcs are ported from net/http fs.go

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length int64
}

func (r httpRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
	//return fmt.Sprintf("bytes %d-%d/*", r.start, r.start+r.length-1)
}

var errNoOverlap = errors.New("invalid range: failed to overlap")

// parseRange parses a Range header string as per RFC 7233.
// errNoOverlap is returned if none of the ranges overlap.
func parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	noOverlap := false
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		start, end, ok := strings.Cut(ra, "-")
		if !ok {
			return nil, errors.New("invalid range")
		}
		start, end = textproto.TrimString(start), textproto.TrimString(end)
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file,
			// and we are dealing with <suffix-length>
			// which has to be a non-negative integer as per
			// RFC 7233 Section 2.1 "Byte-Ranges".
			if end == "" || end[0] == '-' {
				return nil, errors.New("invalid range")
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if i < 0 || err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= size {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				continue
			}
			r.start = i
			if end == "" {
				r.length = sizePerRequst
				if r.length > size-r.start {
					r.length = size - r.start
				}
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, errNoOverlap
	}
	return ranges, nil
}
