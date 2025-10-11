package util

import "context"

type contextKey string

const (
	bucketNameKey    contextKey = "bucket_name"
	correlationIDKey contextKey = "correlation_id"
)

// SetBucketName stores the authenticated bucket name in the context
func SetBucketName(ctx context.Context, bucketName string) context.Context {
	return context.WithValue(ctx, bucketNameKey, bucketName)
}

// GetBucketName retrieves the authenticated bucket name from the context
func GetBucketName(ctx context.Context) (string, bool) {
	bucketName, ok := ctx.Value(bucketNameKey).(string)
	return bucketName, ok
}

// SetCorrelationID stores the correlation ID in the context
func SetCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}
