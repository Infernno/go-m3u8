package main

import (
	"bufio"
	"crypto/aes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/grafov/m3u8"
)

func selectMaxResolution(master m3u8.MasterPlaylist) *m3u8.Variant {
	var variant *m3u8.Variant = nil
	var width int = 0

	for _, v := range master.Variants {
		dimensions := strings.Split(v.Resolution, "x")
		w, err := strconv.Atoi(dimensions[0])

		if err != nil {
			log.Println("Failed to parse resolution:", v.Resolution)
			continue
		}

		if w > width {
			width = w
			variant = v
		}
	}

	if variant == nil {
		log.Fatalln("Variant = nil")
	}

	return variant
}

func getBaseUrl(url string) string {
	end := strings.LastIndex(url, "?")

	if end != -1 {
		for i := end; i >= 0; i-- {
			if url[i] == '/' {
				return url[:i]
			}
		}
	}

	log.Fatalln("Failed to get base url", url)
	return ""
}

func getFileName(url string) string {
	end := strings.LastIndex(url, "?")

	if end == -1 {
		slash := strings.LastIndex(url, "/")
		return url[slash+1:]
	}

	for i := end; i >= 0; i-- {
		if url[i] == '/' {
			return url[i+1 : end]
		}
	}

	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func getPlaylist(url string) (m3u8.Playlist, m3u8.ListType, error) {
	name := getFileName(url)
	playListPath := filepath.Join("downloads", name)

	_, err := GetOrDownload(playListPath, url)

	if err != nil {
		return nil, 0, err
	}

	file, err := os.Open(playListPath)

	if err != nil {
		return nil, 0, err
	}

	defer file.Close()

	playList, listType, err := m3u8.DecodeFrom(bufio.NewReader(file), true)

	if err != nil {
		return nil, 0, err
	}

	return playList, listType, nil
}

func getSegments(playlistPath string) ([]string, error) {
	path, err := filepath.Abs(playlistPath)

	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := make([]string, 10)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "https://") {
			result = append(result, line)
		}
	}

	return result, nil
}

func decryptMessage(key []byte, encryptedData []byte) ([]byte, error) {
	_, err := aes.NewCipher(key)

	if err != nil {
		return nil, fmt.Errorf("could not create new cipher: %v", err)
	}

	//stream, err := cipher.NewCBCDecrypter(block)

	if err != nil {
		return nil, fmt.Errorf("could not create new cipher: %v", err)
	}

	// stream.XORKeyStream(encryptedData, encryptedData)
	return nil, fmt.Errorf("could not create new cipher: %v", err)

	// return cipherText, nil
}

func main2() {
	/*
		m3u8_uri := flag.String("master", "", "URL link to master m3u8")
		flag.Parse()

		if m3u8_uri == nil || len(*m3u8_uri) <= 0 {
			log.Fatalln("Master m3u8 is invalid!")
		}
	*/

	masterUrl := "https://topaz.viacomcbs.digital/h/a/dG9wYXoxYGRiODBiNTdjLWZhMzktMTFlYS04MzRkLTcwZGYyZjg2NmFjZWBiNzBjNzMyYTNiYjcwYmJjNTRmZjUyZWNjOTNlNDFmZDQwNDFmYjRhYGRqeUlaVVBmWE5PeUV2Ni1BMnJKRVk3QllRamZIWVhTSWdsRTlGcXV2X28/master.m3u8?mgid=mgid:arc:episode:shared.smithsonian.us:db80b57c-fa39-11ea-834d-70df2f866ace&ts=1635196429-0&pthash=b70c732a3bb70bbc54ff52ecc93e41fd4041fb4a&hdnts=exp=1702316491~acl=/h/a/dG9wYXoxYGRiODBiNTdjLWZhMzktMTFlYS04MzRkLTcwZGYyZjg2NmFjZWBiNzBjNzMyYTNiYjcwYmJjNTRmZjUyZWNjOTNlNDFmZDQwNDFmYjRhYGRqeUlaVVBmWE5PeUV2Ni1BMnJKRVk3QllRamZIWVhTSWdsRTlGcXV2X28/*~hmac=cb8c35e7e82085584476e648119179e56d33bc9c56f5657e04fa9ccab93effee&CMCD=mtp%3D500%2Cot%3Dm%2Csf%3Dh%2Csid%3D%226697f612-a332-4cca-879e-6c2a804b9af9%22%2Csu"

	playList, _, err := getPlaylist(masterUrl)

	if err != nil {
		log.Fatalln("Failed to get master playlist", err)
	}

	masterPlaylist, ok := playList.(*m3u8.MasterPlaylist)

	if !ok {
		log.Fatalln("Cannot get master playlist")
	}

	variant := selectMaxResolution(*masterPlaylist)

	baseUrl := getBaseUrl(masterUrl)
	mediaPlaylistUrl, err := url.JoinPath(baseUrl, variant.URI)

	if err != nil {
		log.Fatalln("Failed to build media playlist url:", err)
	}

	mediaPlaylistFile, err := GetOrDownload(filepath.Join("downloads", getFileName(mediaPlaylistUrl)), mediaPlaylistUrl)

	if err != nil {
		log.Fatalln("Failed to get media playlist:", err)
	}

	mediaSegments, err := getSegments(mediaPlaylistFile)

	if err != nil {
		log.Fatalln("Failed to get media segments:", err)
	}

	/*

		log.Println("Media playlist url:", mediaPlaylistUrl)

		playList, _, err = getPlaylist(mediaPlaylistUrl)

		if err != nil {
			log.Fatalln("Failed to get media playlist:", err)
		}

		mediaPlaylist := playList.(*m3u8.MediaPlaylist)
	*/

	group := sync.WaitGroup{}

	for _, s := range mediaSegments {
		if s == "" {
			log.Println("Null segment")
			continue
		}

		segmentName := getFileName(s)
		segmentPath := filepath.Join("downloads", segmentName)

		segmentsExists, err := CheckFileExist(segmentPath)

		if err != nil {
			log.Println("Failed to check segment file:", err)
		}

		if segmentsExists {
			log.Println("Skipping", segmentName)
			continue
		}

		log.Println("Downloading", s, "out of", len(mediaSegments))

		go GetOrDownloadParallel(segmentPath, s, &group)
	}

	group.Wait()
}
