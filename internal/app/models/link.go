package models

type LinkInfo struct {
	Long      string
	UUID      string
	IsDeleted bool
}

type LinkJSON struct {
	UUID      string `json:"uuid,omitempty"`
	Short     string `json:"short_url"`
	Long      string `json:"original_url"`
	IsDeleted bool   `json:"is_deleted"`
}
