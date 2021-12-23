package models

type LinkInfo struct {
	Long string
	UUID string
}

type LinkJson struct {
	UUID  string `json:"uuid,omitempty"`
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}
