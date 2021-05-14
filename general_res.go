package bitbank

type GeneralRes struct {
	Status int8        `json:"success"`
	Data   interface{} `json:"data"`
}
