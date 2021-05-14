package bitbank

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Chen33D17017/go-bitbank/model"
)

func GetPrice(cryp string) (model.Price, error) {
	// https://medium.com/@alain.drolet.0/how-to-unmarshal-an-array-of-json-objects-of-different-types-into-a-go-struct-10eab5f9a3a2
	var rst model.PriceRst
	url := fmt.Sprintf("https://public.bitbank.cc/%s_jpy/ticker", cryp)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return rst.Data, fmt.Errorf("Fail to buiild request: %s", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return rst.Data, fmt.Errorf("Request err: %s", err.Error())
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&rst)
	if err != nil {
		return rst.Data, err
	}

	return rst.Data, nil
}
