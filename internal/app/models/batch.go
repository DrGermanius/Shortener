package models

type BatchOriginal struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShort struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
