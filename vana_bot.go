package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Vana struct {
	gameURL string
}

func NewVana() *Vana {
	return &Vana{
		gameURL: "https://www.vanadatahero.com/api",
	}
}

func countdown(seconds int) {
	for seconds > 0 {
		mins := seconds / 60
		secs := seconds % 60
		fmt.Printf("  ⏱  Waiting.. %02d:%02d seconds\r", mins, secs)
		time.Sleep(1 * time.Second)
		seconds--
	}
	fmt.Println("")
}

func (v *Vana) commonHeader(query string) map[string]string {
	return map[string]string{
		"Host":                         "www.vanadatahero.com",
		"content-type":                 "application/json",
		"user-agent":                   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0",
		"accept":                       "*/*",
		"x-telegram-web-app-init-data": query,
		"sec-fetch-site":               "same-origin",
		"sec-fetch-mode":               "cors",
		"sec-fetch-dest":               "empty",
		"referer":                      "https://www.vanadatahero.com/challenges",
		"accept-language":              "en-US,en;q=0.9",
		"priority":                     "u=1, i",
	}
}

func (v *Vana) sendRequest(method, url string, headers map[string]string, jsonData map[string]interface{}) ([]byte, error) {
	client := &http.Client{}
	var reqBody []byte
	if jsonData != nil {
		reqBody, _ = json.Marshal(jsonData)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		if strings.Contains(url, "/tasks/1") {
			return []byte("Ok"), nil
		}
		return body, nil
	}

	return nil, fmt.Errorf("error: %s", body)
}

func (v *Vana) postRequest(url string, headers map[string]string, jsonData map[string]interface{}) ([]byte, error) {
	return v.sendRequest("POST", url, headers, jsonData)
}

func (v *Vana) getRequest(url string, headers map[string]string) ([]byte, error) {
	return v.sendRequest("GET", url, headers, nil)
}

func (v *Vana) getPlayer(sessionHeaders map[string]string) (string, error) {
	url := v.gameURL + "/player"
	playerInfo, err := v.getRequest(url, sessionHeaders)
	if err != nil {
		return "", err
	}

	var playerData map[string]interface{}
	if err := json.Unmarshal(playerInfo, &playerData); err == nil {
		fmt.Printf("\n  😗 Player: %s - Total Point: %v🌸\n", playerData["tgUsername"], playerData["points"])
		return playerData["tgUsername"].(string), nil
	}
	if strings.Contains(string(playerInfo), "expired") {
		return "expired", nil
	}
	return "", fmt.Errorf("could not get player info")
}

func (v *Vana) play(sessionHeaders map[string]string, point float64, playerName string) {
	url := v.gameURL + "/tasks/1"
	jsonData := map[string]interface{}{
		"status": "completed",
		"points": point,
	}

	playInfo, err := v.postRequest(url, sessionHeaders, jsonData)
	if err == nil {
		fmt.Printf("  🎲 Player: %s - Play: %s: +%.1f Points 🌸\n", playerName, playInfo, point)
	} else {
		fmt.Printf("  ❌ Player: %s - Play: %v\n", playerName, err)
	}
}

func (v *Vana) start(query string, total int) {
	if query == "" {
		fmt.Println("❌ QUERY_TOKEN IS EMPTY.")
		return
	}

	sessionHeaders := v.commonHeader(query)
	playerName, err := v.getPlayer(sessionHeaders)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	if playerName == "expired" {
		fmt.Println("\n❌ QUERY_TOKEN EXPIRED. PLEASE RELOAD THE GAME AND GET A NEW TOKEN.")
		return
	}

	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Minute) // Ticker for logging every minute
	defer ticker.Stop()

	for i := 0; i < total; i++ {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime)
			fmt.Printf("⏰ Running time: %v\n", elapsed)
		default:
			fmt.Printf("\n  ➡️  Round: %d/%d\n", i+1, total)
			point := 40 + rand.Float64()*(20-10)
			v.play(sessionHeaders, point, playerName)
			countdown(20)
			fmt.Println()

			// Check if token has expired
			playerName, err := v.getPlayer(sessionHeaders)
			if err != nil {
				fmt.Println("ERROR:", err)
				return
			}

			if playerName == "expired" {
				fmt.Println("\n❌ QUERY_TOKEN EXPIRED. PLEASE RELOAD THE GAME AND GET A NEW TOKEN.")
				return
			}
		}
	}

	v.getPlayer(sessionHeaders)
}

func main() {

	fmt.Print("\n🌸 ENTER QUERY_TOKEN: \n    ➡️   ")

	// Sử dụng bufio để nhập token đầy đủ
	reader := bufio.NewReader(os.Stdin)
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query) // Bỏ dấu newline hoặc khoảng trắng

	fmt.Println("    - OK\n")

	var total int
	fmt.Print("\n🌸 ENTER THE NUMBER OF TIMES TO PLAY: \n    ➡️   ")
	fmt.Scan(&total)

	fmt.Println("    - OK")

	vanaBot := NewVana()
	vanaBot.start(query, total)
}
