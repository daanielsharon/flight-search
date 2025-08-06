package sharedmodels

type SearchRequest struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Date       string `json:"date"`
	Passengers int    `json:"passengers"`
}
