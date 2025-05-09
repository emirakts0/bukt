package request

type CreateKVRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   int64  `json:"ttl"` // in seconds
}
