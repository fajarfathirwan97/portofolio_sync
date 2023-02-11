package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"sync"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
)

func getPriceStock(symbol string) interface{} {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://data.alpaca.markets/v2/stocks/%v/trades", symbol), nil)
	req.Header.Set("APCA-API-KEY-ID", "PKVLMI3ZCTI26MDA7Z15")
	req.Header.Set("APCA-API-SECRET-KEY", "G8v2sUXFL7Tx3Ryr7Fs0BHZB70Cmltq30yz1jCk9")
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer res.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(res.Body)
	var b interface{}
	json.Unmarshal(bodyBytes, &b)
	return b

}
func main() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := config.Client(context.Background())

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	spreadsheetId := "17oKav5iiLcZ2ZEAYvSl0c4di3-p48V3lj4KF8W8NLSg"
	readRange := "Porto"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	wg := &sync.WaitGroup{}
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for index, row := range resp.Values {
			if index == 0 {
				continue
			}
			wg.Add(1)
			go updateSheet(row, srv, spreadsheetId, index, wg)
		}
	}
	wg.Wait()
	log.Println("SYNC COMPLETED")
}

func updateSheet(row []interface{}, srv *sheets.Service, spreadsheetId string, index int, wg *sync.WaitGroup) {
	res := getPriceStock(row[0].(string)).(map[string]interface{})
	trades := res["trades"].([]interface{})[0].(map[string]interface{})
	m := trades["p"]
	upd := srv.Spreadsheets.Values.Update(spreadsheetId,
		fmt.Sprintf("Porto!F%v", index+1),
		&sheets.ValueRange{
			Values: [][]interface{}{{m}},
		})
	upd.ValueInputOption("RAW")
	upd.Do()
	wg.Done()
}
