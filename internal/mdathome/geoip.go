package mdathome

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

var geodb *geoip2.Reader

func prepareGeoIPDatabase() {

	// Set MaxMind database filename
	maxMindDatabaseFilename := "GeoLite2-Country.mmdb"

	// Check if database already downloaded
	if _, err := os.Stat(maxMindDatabaseFilename); os.IsNotExist(err) {
		// Log
		log.Warnf("Downloding geolocation data in the background...")

		// Prepare URL
		databaseURL := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=%s&suffix=tar.gz", clientSettings.MaxMindLicenseKey)

		// Download database if not exists
		resp, err := http.Get(databaseURL)
		if err != nil {
			log.Fatalf("Failed to download MaxMind database: %v", err)
		}
		defer resp.Body.Close()

		// Uncompress archive
		uncompressedArchive, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatal("Failed to uncompress MaxMind database: %v", err)
		}
		defer uncompressedArchive.Close()

		// Loop through tar archive entries
		tarReader := tar.NewReader(uncompressedArchive)
		for {
			// Get next tar archive entry
			header, err := tarReader.Next()

			// If finished entire file and no database found
			if err == io.EOF {
				log.Fatalln("Unable to find MaxMind database file in archive")
			}

			// If EOF
			if err != nil {
				log.Fatalf("Failed to extract MaxMind database file: %v", err)
			}

			// If tar archive entry matches our requirements, save to file
			if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, maxMindDatabaseFilename) {
				outFile, err := os.Create(maxMindDatabaseFilename)
				if err != nil {
					log.Fatalf("Failed to create MaxMind database file: %s", err.Error())
				}
				defer outFile.Close()

				if _, err := io.Copy(outFile, tarReader); err != nil {
					log.Fatalf("Failed to write to MaxMind database file: %s", err.Error())
				}

				log.Warnf("Downloaded MaxMind database")
				break
			}
		}
	}

	// Open MaxMind database
	var err error
	geodb, err = geoip2.Open(maxMindDatabaseFilename)
	if err != nil {
		log.Fatalf("Unable to open database %s for geolocation: %v", maxMindDatabaseFilename, err)
	}
	log.Warnf("Loaded geolocation database")

}
