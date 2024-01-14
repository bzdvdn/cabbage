package cabbage

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// RedisBroker is cabbage broker for redis
type RedisBroker struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisBroker creates with given redis connection with context
func NewRedisBrokerWithContext(ctx context.Context, url string) (*RedisBroker, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return &RedisBroker{
		client: client,
		ctx:    ctx,
	}, nil
}

// NewRedisBroker creates with given redis connection
func NewRedisBroker(url string) (*RedisBroker, error) {
	ctx := context.Background()
	return NewRedisBrokerWithContext(ctx, url)
}

// EnableQueueForWorker ...
func (b *RedisBroker) EnableQueueForWorker(queueName string) error {
	return nil
}

// Close redis broker
func (b *RedisBroker) Close() {
	b.client.Close()
}

// SendCabbageMessage send cabbage message to redis broker
func (b *RedisBroker) SendCabbageMessage(queueName string, cbMessage *CabbageMessage) error {
	js, err := json.Marshal(cbMessage)
	if err != nil {
		return err
	}
	err = b.client.LPush(b.ctx, queueName, string(js)).Err()
	return err
}

// GetCabbageMessage get cabbage message from redis broker
func (b *RedisBroker) GetCabbageMessage(queueName string) (*CabbageMessage, error) {
	item, err := b.client.LPop(b.ctx, queueName).Result()
	if err != nil {
		return nil, err
	}
	var cbMessage CabbageMessage
	err = json.Unmarshal([]byte(item), &cbMessage)
	if err != nil {
		return nil, err
	}
	return &cbMessage, nil
}
