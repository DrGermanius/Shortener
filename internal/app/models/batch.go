package models

type BatchOriginal struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShort struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
