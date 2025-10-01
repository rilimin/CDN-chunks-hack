package utils

import (
	"io"
	"time"
)

type (
	ProgressReader struct {
		io.Reader
		Upload *Upload
	}

	FileInfo struct {
		FileName string
		Size     string
		Links    []string
	}

	DiscordWebhook struct {
		Type         int           `json:"type"`
		Content      string        `json:"content"`
		Mentions     []interface{} `json:"mentions"`
		MentionRoles []interface{} `json:"mention_roles"`
		Attachments  []struct {
			ID                 string `json:"id"`
			Filename           string `json:"filename"`
			Size               int    `json:"size"`
			URL                string `json:"url"`
			ProxyURL           string `json:"proxy_url"`
			ContentType        string `json:"content_type"`
			ContentScanVersion int    `json:"content_scan_version"`
		} `json:"attachments"`
		Embeds          []interface{} `json:"embeds"`
		Timestamp       time.Time     `json:"timestamp"`
		EditedTimestamp interface{}   `json:"edited_timestamp"`
		Flags           int           `json:"flags"`
		Components      []interface{} `json:"components"`
		ID              string        `json:"id"`
		ChannelID       string        `json:"channel_id"`
		Author          struct {
			ID            string      `json:"id"`
			Username      string      `json:"username"`
			Avatar        interface{} `json:"avatar"`
			Discriminator string      `json:"discriminator"`
			PublicFlags   int         `json:"public_flags"`
			Flags         int         `json:"flags"`
			Bot           bool        `json:"bot"`
			GlobalName    interface{} `json:"global_name"`
			Clan          interface{} `json:"clan"`
			PrimaryGuild  interface{} `json:"primary_guild"`
		} `json:"author"`
		Pinned          bool   `json:"pinned"`
		MentionEveryone bool   `json:"mention_everyone"`
		Tts             bool   `json:"tts"`
		WebhookID       string `json:"webhook_id"`
	}

	RefreshRequest struct {
		AttachmentURLs []string `json:"attachment_urls"`
	}

	RefreshResponse struct {
		RefreshedUrls []struct {
			Original  string `json:"original"`
			Refreshed string `json:"refreshed"`
		} `json:"refreshed_urls"`
	}

	Upload struct {
		TotalSize    int64
		UploadedSize int64
		StartTime    time.Time
	}
)
