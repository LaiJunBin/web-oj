package mongodb

import (
	"context"
	"fmt"
	"time"

	sharedtypes "judge-backend/internal/app/shared/types"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	maxRetries    = 3
	retryBaseWait = 500 * time.Millisecond
)

// isTransientError reports whether err is a transient infrastructure failure
// that is safe to retry: DNS resolution failures, brief network blips, and
// server-selection timeouts all fall into this category.
func isTransientError(err error) bool {
	return err != nil && (mongo.IsTimeout(err) || mongo.IsNetworkError(err))
}

func withRetry(logger sharedtypes.Logger, tag string, op func() error) error {
	wait := retryBaseWait
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := op()
		if err == nil {
			return nil
		}
		if !isTransientError(err) || attempt == maxRetries {
			return err
		}
		logger.Warnf("%s transient error (attempt %d/%d), retrying in %s: %v", tag, attempt+1, maxRetries, wait, err)
		time.Sleep(wait)
		wait *= 2
	}
	return nil
}

// RetryCollection wraps *mongo.Collection and implements sharedtypes.MongoCollection.
// It transparently retries operations that fail with transient errors
// (DNS hiccups, brief network blips, server-selection timeouts).
// Permanent errors (duplicate-key, validation, ErrNoDocuments) are returned
// immediately without retry.
//
// *mongo.Collection is not a persistent connection — it is a lightweight
// namespace reference (client → database → collection). The driver's internal
// connection pool manages actual TCP connections, so retrying with the same
// instance gives the driver a fresh chance to select a healthy server.
type RetryCollection struct {
	col    *mongo.Collection
	logger sharedtypes.Logger
	name   string // collection name, used in retry log messages
}

// Ensure RetryCollection satisfies the interface at compile time.
var _ sharedtypes.MongoCollection = (*RetryCollection)(nil)

func NewRetryCollection(col *mongo.Collection, logger sharedtypes.Logger) *RetryCollection {
	return &RetryCollection{col: col, logger: logger, name: col.Name()}
}

func (r *RetryCollection) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var result *mongo.UpdateResult
	err := withRetry(r.logger, fmt.Sprintf("[%s] UpdateOne", r.name), func() error {
		var e error
		result, e = r.col.UpdateOne(ctx, filter, update, opts...)
		return e
	})
	return result, err
}

func (r *RetryCollection) UpdateMany(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var result *mongo.UpdateResult
	err := withRetry(r.logger, fmt.Sprintf("[%s] UpdateMany", r.name), func() error {
		var e error
		result, e = r.col.UpdateMany(ctx, filter, update, opts...)
		return e
	})
	return result, err
}

func (r *RetryCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	var cursor *mongo.Cursor
	err := withRetry(r.logger, fmt.Sprintf("[%s] Find", r.name), func() error {
		var e error
		cursor, e = r.col.Find(ctx, filter, opts...)
		return e
	})
	return cursor, err
}

func (r *RetryCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	var count int64
	err := withRetry(r.logger, fmt.Sprintf("[%s] CountDocuments", r.name), func() error {
		var e error
		count, e = r.col.CountDocuments(ctx, filter, opts...)
		return e
	})
	return count, err
}

func (r *RetryCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	var cursor *mongo.Cursor
	err := withRetry(r.logger, fmt.Sprintf("[%s] Aggregate", r.name), func() error {
		var e error
		cursor, e = r.col.Aggregate(ctx, pipeline, opts...)
		return e
	})
	return cursor, err
}

// FindOne eagerly executes the query so that retry can occur before returning.
// The caller uses .Err() / .Decode() on the returned *mongo.SingleResult
// exactly as before — the retry is invisible to them.
func (r *RetryCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	var sr *mongo.SingleResult
	for attempt := 0; attempt <= maxRetries; attempt++ {
		sr = r.col.FindOne(ctx, filter, opts...)
		err := sr.Err()
		if err == nil || !isTransientError(err) {
			break
		}
		if attempt < maxRetries {
			wait := retryBaseWait * time.Duration(1<<attempt)
			r.logger.Warnf("[%s] FindOne transient error (attempt %d/%d), retrying in %s: %v",
				r.name, attempt+1, maxRetries, wait, err)
			time.Sleep(wait)
		}
	}
	return sr
}
