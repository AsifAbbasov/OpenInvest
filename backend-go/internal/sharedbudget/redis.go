package sharedbudget

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const operationTimeout = time.Second

var ErrUnavailable = errors.New("shared budget backend unavailable")

type Authority interface {
	Ping(context.Context) error
	AdmitEndpoint(context.Context, string, int, int, time.Duration) (bool, error)
	AdmitProvider(context.Context, int, time.Duration) (bool, error)
	ObserveProviderAllowance(context.Context, int, time.Duration) error
	Close() error
}

type RedisAuthority struct {
	client    *redis.Client
	namespace string
}

var endpointAdmissionScript = redis.NewScript(`
local global = tonumber(redis.call("GET", KEYS[1]) or "0")
local client = tonumber(redis.call("GET", KEYS[2]) or "0")
local global_limit = tonumber(ARGV[1])
local client_limit = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])
if global >= global_limit or client >= client_limit then
	return 0
end
global = redis.call("INCR", KEYS[1])
if global == 1 then
	redis.call("PEXPIRE", KEYS[1], ttl)
end
client = redis.call("INCR", KEYS[2])
if client == 1 then
	redis.call("PEXPIRE", KEYS[2], ttl)
end
return 1
`)

var providerAdmissionScript = redis.NewScript(`
local used = tonumber(redis.call("GET", KEYS[1]) or "0")
local limit = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])
if used >= limit then
	return 0
end
local remote = redis.call("GET", KEYS[2])
if remote and tonumber(remote) <= 0 then
	return 0
end
used = redis.call("INCR", KEYS[1])
if used == 1 then
	redis.call("PEXPIRE", KEYS[1], ttl)
end
if remote then
	redis.call("DECR", KEYS[2])
end
return 1
`)

var providerObservationScript = redis.NewScript(`
local remaining = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])
local current = redis.call("GET", KEYS[1])
if not current then
	redis.call("SET", KEYS[1], remaining, "PX", ttl)
	return 1
end
if remaining < tonumber(current) then
	redis.call("SET", KEYS[1], remaining, "KEEPTTL")
end
local current_ttl = redis.call("PTTL", KEYS[1])
if current_ttl < ttl then
	redis.call("PEXPIRE", KEYS[1], ttl)
end
return 1
`)

func NewRedisAuthority(rawURL, namespace string) (*RedisAuthority, error) {
	rawURL = strings.TrimSpace(rawURL)
	namespace = strings.TrimSpace(namespace)
	if rawURL == "" {
		return nil, errors.New("shared budget Redis URL is required")
	}
	if namespace == "" || strings.ContainsAny(namespace, " \t\r\n") {
		return nil, errors.New("shared budget namespace is invalid")
	}
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse shared budget Redis URL: %w", err)
	}
	options.MaxRetries = 0
	options.DialTimeout = operationTimeout
	options.ReadTimeout = operationTimeout
	options.WriteTimeout = operationTimeout
	return &RedisAuthority{
		client:    redis.NewClient(options),
		namespace: namespace,
	}, nil
}

func (authority *RedisAuthority) Ping(ctx context.Context) error {
	if authority == nil || authority.client == nil {
		return ErrUnavailable
	}
	ctx, cancel := authority.operationContext(ctx)
	defer cancel()
	if err := authority.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (authority *RedisAuthority) AdmitEndpoint(
	ctx context.Context,
	clientIdentity string,
	clientLimit int,
	globalLimit int,
	window time.Duration,
) (bool, error) {
	if authority == nil || authority.client == nil {
		return false, ErrUnavailable
	}
	if strings.TrimSpace(clientIdentity) == "" || clientLimit <= 0 || globalLimit <= 0 || window <= 0 {
		return false, errors.New("shared endpoint budget configuration is invalid")
	}
	sum := sha256.Sum256([]byte(clientIdentity))
	clientKey := authority.key("corporate-actions:endpoint:client:" + hex.EncodeToString(sum[:]))
	globalKey := authority.key("corporate-actions:endpoint:global")
	ctx, cancel := authority.operationContext(ctx)
	defer cancel()
	result, err := endpointAdmissionScript.Run(
		ctx,
		authority.client,
		[]string{globalKey, clientKey},
		globalLimit,
		clientLimit,
		window.Milliseconds(),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("%w: endpoint admission: %v", ErrUnavailable, err)
	}
	return result == 1, nil
}

func (authority *RedisAuthority) AdmitProvider(ctx context.Context, limit int, window time.Duration) (bool, error) {
	if authority == nil || authority.client == nil {
		return false, ErrUnavailable
	}
	if limit <= 0 || window <= 0 {
		return false, errors.New("shared provider budget configuration is invalid")
	}
	ctx, cancel := authority.operationContext(ctx)
	defer cancel()
	result, err := providerAdmissionScript.Run(
		ctx,
		authority.client,
		[]string{
			authority.key("tinvest:provider:used"),
			authority.key("tinvest:provider:remote-remaining"),
		},
		limit,
		window.Milliseconds(),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("%w: provider admission: %v", ErrUnavailable, err)
	}
	return result == 1, nil
}

func (authority *RedisAuthority) ObserveProviderAllowance(ctx context.Context, remaining int, resetAfter time.Duration) error {
	if authority == nil || authority.client == nil {
		return ErrUnavailable
	}
	if remaining < 0 || resetAfter <= 0 || resetAfter > time.Hour {
		return errors.New("provider allowance observation is invalid")
	}
	ctx, cancel := authority.operationContext(ctx)
	defer cancel()
	if err := providerObservationScript.Run(
		ctx,
		authority.client,
		[]string{authority.key("tinvest:provider:remote-remaining")},
		remaining,
		resetAfter.Milliseconds(),
	).Err(); err != nil {
		return fmt.Errorf("%w: provider allowance observation: %v", ErrUnavailable, err)
	}
	return nil
}

func (authority *RedisAuthority) Close() error {
	if authority == nil || authority.client == nil {
		return nil
	}
	return authority.client.Close()
}

func (authority *RedisAuthority) key(suffix string) string {
	return authority.namespace + ":" + suffix
}

func (authority *RedisAuthority) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if deadline, ok := parent.Deadline(); ok && time.Until(deadline) <= operationTimeout {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, operationTimeout)
}
