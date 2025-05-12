package request

type CreateKVRequest struct {
	Key        string `json:"key" binding:"required,min=1,max=255"`
	Value      string `json:"value" binding:"required,min=1"`
	TTL        int64  `json:"ttl" binding:"required,gt=0"` // in seconds
	SingleRead bool   `json:"single_read" binding:"required"`
}
