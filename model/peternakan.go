package model

type Peternakan struct {
	User_id            int64   `json:"user_id"`
	Nama_peternakan    string  `json:"nama_peternakan"`
	Street             string  `json:"street"`
	City               string  `json:"city"`
	State              string  `json:"state"`
	PostalCode         string  `json:"postal_code"`
	Country            string  `json:"country"`
	Lat                float64 `json:"lat"`
	Lon                float64 `json:"lon"`
	PeternakanImageURL string  `json:"image_farm"`
}
