package platforms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	state "main/internal/core/models"
)

const PlatformCobalt state.PlatformName = "Cobalt"

type CobaltPlatform struct {
	name   state.PlatformName
	apiURL string
}

func init() {
	apiURL := os.Getenv("COBALT_API_URL")
	if apiURL == "" {
		return // COBALT_API_URL nahi hai toh register mat karo
	}
	Register(75, &CobaltPlatform{
		name:   PlatformCobalt,
		apiURL: apiURL,
	})
}

func (c *CobaltPlatform) Name() state.PlatformName {
	return c.name
}

// Cobalt sirf download karta hai, search nahi
func (c *CobaltPlatform) CanSearch() bool {
	return false
}

func (c *CobaltPlatform) Search(query string, video bool) ([]*state.Track, error) {
	return nil, fmt.Errorf("cobalt does not support search")
}

// Sirf YouTube URLs handle karega
func (c *CobaltPlatform) CanGetTracks(query string) bool {
	return false
}

func (c *CobaltPlatform) GetTracks(query string, video bool) ([]*state.Track, error) {
	return nil, fmt.Errorf("cobalt does not support GetTracks")
}

// YouTube platform ke tracks download kar sakta hai
func (c *CobaltPlatform) CanDownload(source state.PlatformName) bool {
	return source == "YouTube" || source == "Spotify" || source == "SoundCloud"
}

func (c *CobaltPlatform) Download(
	ctx context.Context,
	track *state.Track,
	mystic *telegram.NewMessage,
) (string, error) {
	// Cobalt API request
	mode := "audio"
	if track.Video {
		mode = "auto"
	}

	reqBody := map[string]interface{}{
		"url":          track.URL,
		"downloadMode": mode,
		"audioFormat":  "mp3",
		"videoQuality": "720",
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var cobaltResp struct {
		Status   string `json:"status"`
		URL      string `json:"url"`
		Filename string `json:"filename"`
	}
	json.NewDecoder(resp.Body).Decode(&cobaltResp)

	if cobaltResp.Status == "error" || cobaltResp.URL == "" {
		return "", fmt.Errorf("cobalt: failed to get download URL")
	}

	// File download karo
	ext := "mp3"
	if track.Video {
		ext = "mp4"
	}
	fileName := fmt.Sprintf("downloads/%s.%s", track.ID, ext)
	os.MkdirAll(filepath.Dir(fileName), 0755)

	fileResp, err := http.Get(cobaltResp.URL)
	if err != nil {
		return "", err
	}
	defer fileResp.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, fileResp.Body)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func (c *CobaltPlatform) CanGetRecommendations() bool {
	return false
}

func (c *CobaltPlatform) GetRecommendations(track *state.Track) ([]*state.Track, error) {
	return nil, fmt.Errorf("cobalt does not support recommendations")
}

// Helper
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
