package scanners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type ImportList struct {
	ID      int                    `json:"id"`
	Name    string                 `json:"name"`
	Enabled bool                   `json:"enabled"`
	Raw     map[string]interface{} `json:"-"`
}

type Scanners struct {
	SonarrLists []ImportList
	RadarrLists []ImportList
}

var (
	sonarrURL string
	sonarrKey string

	radarrURL string
	radarrKey string

	scanners Scanners
)

func getImportLists(apiUrl, apiKey string) ([]ImportList, error) {
	req, _ := http.NewRequest("GET", apiUrl+"/api/v3/importlist", nil)
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var raw []map[string]interface{}
	json.Unmarshal(body, &raw)

	var lists []ImportList
	for _, item := range raw {
		list := ImportList{
			ID:      int(item["id"].(float64)),
			Name:    item["name"].(string),
			Enabled: item["enabled"].(bool),
			Raw:     item,
		}

		lists = append(lists, list)
	}

	return lists, nil
}

func updateImportList(apiUrl, apiKey string, list ImportList, enable bool) error {
	list.Raw["enabled"] = enable
	payload, _ := json.Marshal(list.Raw)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v3/importlist/%d", apiUrl, list.ID), bytes.NewReader(payload))
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Radarr error: %s", string(body))
	}

	return nil
}

func refreshList(apiUrl, apiKey string, list ImportList) error {
	var err error
	err = updateImportList(apiUrl, apiKey, list, false)
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return updateImportList(apiUrl, apiKey, list, true)
}

func init() {
	godotenv.Load()

	sonarrURL = os.Getenv("SONARR_URL")
	sonarrKey = os.Getenv("SONARR_API_KEY")

	radarrURL = os.Getenv("RADARR_URL")
	radarrKey = os.Getenv("RADARR_API_KEY")

	if sonarrURL == "" || sonarrKey == "" || radarrURL == "" || radarrKey == "" {
		log.Fatal("Les variables d'environnement SONARR_URL, SONARR_API_KEY, RADARR_URL et RADARR_API_KEY sont requises.")
	}

	var err error
	scanners.SonarrLists, err = getImportLists(sonarrURL, sonarrKey)
	if err != nil {
		log.Fatalf("Erreur récupération Sonarr: %v", err)
	}

	scanners.RadarrLists, err = getImportLists(radarrURL, radarrKey)
	if err != nil {
		log.Fatalf("Erreur récupération Radarr: %v", err)
	}
}

func RefreshSonarr() {
	for _, list := range scanners.SonarrLists {
		err := refreshList(sonarrURL, sonarrKey, list)
		if err != nil {
			log.Println("Erreur mise à jour Sonarr:", err)
		}
	}
}

func RefreshRadarr() {
	for _, list := range scanners.RadarrLists {
		err := refreshList(radarrURL, radarrKey, list)
		if err != nil {
			log.Println("Erreur mise à jour Radarr:", err)
		}
	}
}
