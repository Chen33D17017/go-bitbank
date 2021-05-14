package bitbank

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func readUTC(timestamp int64) string {
	return time.Unix(timestamp/1000, 0).Format("2006-01-02")
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func encode(s Secret, content string) string {
	h := hmac.New(sha256.New, []byte(s.ApiSecret))
	h.Write([]byte(content))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

func addHeader(req *http.Request, s Secret, content string) {
	nonce := fmt.Sprint(makeTimestamp())
	req.Header.Add("ACCESS-KEY", s.ApiKey)
	req.Header.Add("ACCESS-NONCE", nonce)
	req.Header.Add("ACCESS-SIGNATURE", encode(s, nonce+content))
	req.Header.Add("Content-Type", "application/json")
}

func apiRequest(req *http.Request, response interface{}) error {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Fail to request: %s", err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Fail to read response body: %s", err.Error())
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}

func getRequest(s Secret, query string, response interface{}) error {
	url := fmt.Sprintf("https://api.bitbank.cc%s", query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Fail to build GET request: %s", err.Error())
	}

	addHeader(req, s, query)
	err = apiRequest(req, response)
	if err != nil {
		return err
	}
	return nil
}

func postRequest(s Secret, endpoint string, payload []byte, response interface{}) error {
	url := fmt.Sprintf("https://api.bitbank.cc%s", endpoint)
	payloadReader := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", url, payloadReader)
	if err != nil {
		return fmt.Errorf("Fail to build POST request: %s", err)
	}

	addHeader(req, s, string(payload))
	err = apiRequest(req, response)
	if err != nil {
		return err
	}
	return nil
}

func CheckAssets(s Secret) ([]Asset, error) {
	var response AssetRst
	err := getRequest(s, "/v1/user/assets", &response)
	if err != nil {
		return response.Data.Assets, err
	}
	return response.Data.Assets, nil
}

func MakeTrade(s Secret, assetType string, buy_sell string, amount float64) (Order, error) {
	url := fmt.Sprintf("/v1/user/spot/order")
	var response OrderRst

	order := OrderRequest{
		Pair:   fmt.Sprintf("%s_jpy", assetType),
		Amount: fmt.Sprintf("%.4f", amount),
		Side:   buy_sell,
		Type:   "market",
	}

	reqBody, _ := json.Marshal(order)
	err := postRequest(s, url, reqBody, &response)
	if err != nil {
		return response.Data, err
	}

	return response.Data, nil
}

func BuyWithJPY(s Secret, assetType string, JPY int64) (Order, error) {
	cryptmsg, err := GetPrice(assetType)
	if err != nil {
		fmt.Println(err.Error())
	}
	cryptPrice, _ := strconv.Atoi(cryptmsg.Buy)
	amount := float64(JPY) / float64(cryptPrice)

	return MakeTrade(s, assetType, "buy", amount)
}

func SellToJPY(s Secret, assetType string, amount float64) (Order, error) {
	return MakeTrade(s, assetType, "sell", amount)
}

func GetTradeHistory(s Secret, assetType string) ([]Trade, error) {
	var response TradeRst
	url := fmt.Sprintf("/v1/user/spot/trade_history?pair=%s_jpy", assetType)
	err := getRequest(s, url, &response)
	if err != nil {
		return nil, err
	}
	return response.Data.Trades, nil
}

func GetOrderInfo(s Secret, assetType, order_id string) (Order, error) {
	var response OrderRst
	url := fmt.Sprintf("/v1/user/spot/order?pair=%s_jpy&order_id=%s", assetType, order_id)
	err := getRequest(s, url, &response)
	if err != nil {
		return response.Data, err
	}
	return response.Data, nil
}
