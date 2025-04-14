package main

import (
	"AlanSnt/PlexRssSync/scanners"
	"crypto/sha256"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.etcd.io/bbolt"
)

var (
	rssURLs    []string
	prevHashes map[string][32]byte

	refreshRate int
	dbPath      = "./db/hash.db"
)

func init() {
	godotenv.Load()

	rssURLs = strings.Split(os.Getenv("PLEX_RSS_URLS"), ",")
	refreshRate = getEnvAsInt("REFRESH_RATE", 2)
	prevHashes = make(map[string][32]byte)

	if len(rssURLs) == 0 {
		log.Fatal("La variable d'environnement RSS_URLS doit contenir au moins une URL.")
	}
}

func getEnvAsInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return result
}

func fetchHash(url string) ([32]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return [32]byte{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return [32]byte{}, err
	}

	return sha256.Sum256(body), nil
}

func triggerScan(apiURL, apiKey string) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		log.Println("Erreur création requête:", err)
		return
	}
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(`{"name":"RescanSeries"}`))

	_, err = client.Do(req)
	if err != nil {
		log.Println("Erreur requête Sonarr/Radarr:", err)
	}
}

func storeHash(url string, hash [32]byte) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("hashes"))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(url), hash[:])
	})
	if err != nil {
		log.Fatal(err)
	}
}

func loadHashes() {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("hashes"))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var hash [32]byte
			copy(hash[:], v)
			prevHashes[string(k)] = hash
			return nil
		})
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	loadHashes()
	ticker := time.NewTicker(time.Duration(refreshRate) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Vérification des flux RSS...")
		for _, url := range rssURLs {
			newHash, err := fetchHash(url)
			if err != nil {
				log.Println("Erreur récupération RSS:", err)
				continue
			}

			if prevHashes[url] != newHash {
				log.Printf("Changement détecté sur %s\n", url)
				prevHashes[url] = newHash
				storeHash(url, newHash)

				go scanners.RefreshRadarr()
				go scanners.RefreshSonarr()
			}
		}
	}
}
