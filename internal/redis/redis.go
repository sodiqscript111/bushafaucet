package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	FaucetStream	= "faucet_jobs"

	ConsumerGroup	= "faucet_workers"

	LockPrefix	= "faucet:lock:"

	LockTTL	= 10 * time.Second
)

type Client struct {
	rdb *redis.Client
}

func NewClient(addr string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:		addr,
		DB:		0,
		DialTimeout:	5 * time.Second,
		ReadTimeout:	5 * time.Second,
		WriteTimeout:	5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) AcquireLock(ctx context.Context, walletAddress string) (bool, error) {
	key := LockPrefix + walletAddress
	ok, err := c.rdb.SetNX(ctx, key, "locked", LockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("acquire lock: %w", err)
	}
	return ok, nil
}

func (c *Client) ReleaseLock(ctx context.Context, walletAddress string) error {
	key := LockPrefix + walletAddress
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("release lock: %w", err)
	}
	return nil
}

func (c *Client) PublishJob(ctx context.Context, claimID, walletAddress, blockchain string) error {
	_, err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream:	FaucetStream,
		Values: map[string]interface{}{
			"claim_id":		claimID,
			"wallet_address":	walletAddress,
			"blockchain":		blockchain,
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("publish job: %w", err)
	}
	slog.Info("published faucet job", "claim_id", claimID, "wallet", walletAddress, "blockchain", blockchain)
	return nil
}

func (c *Client) CreateConsumerGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, FaucetStream, ConsumerGroup, "0").Err()
	if err != nil {

		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			slog.Info("consumer group already exists", "group", ConsumerGroup)
			return nil
		}
		return fmt.Errorf("create consumer group: %w", err)
	}
	slog.Info("created consumer group", "group", ConsumerGroup, "stream", FaucetStream)
	return nil
}

func (c *Client) ReadJobs(ctx context.Context, consumer string, count int64, block time.Duration) ([]redis.XStream, error) {
	result, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:		ConsumerGroup,
		Consumer:	consumer,
		Streams:	[]string{FaucetStream, ">"},
		Count:		count,
		Block:		block,
	}).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read jobs: %w", err)
	}
	return result, nil
}

func (c *Client) AckJob(ctx context.Context, messageID string) error {
	if err := c.rdb.XAck(ctx, FaucetStream, ConsumerGroup, messageID).Err(); err != nil {
		return fmt.Errorf("ack job: %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
