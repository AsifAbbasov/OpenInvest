package sharedbudget

import (
	"context"
	"crypto/rand"
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

// Legacy string counters lack timestamps. Conservatively reserve a full
// rolling allowance for one complete window when converting their key type.
const legacyReservationScript = `
local function reserve_legacy(key, limit, now, ttl)
 redis.call("DEL", key)
 for i = 1, limit do
  redis.call("ZADD", key, now, "legacy-reservation:" .. i)
 end
 redis.call("PEXPIRE", key, ttl)
end
`

var endpointAdmissionScript = redis.NewScript(legacyReservationScript + `
local global_type = redis.call("TYPE", KEYS[1]).ok
local client_type = redis.call("TYPE", KEYS[2]).ok
if (global_type ~= "none" and global_type ~= "zset" and global_type ~= "string") or
   (client_type ~= "none" and client_type ~= "zset" and client_type ~= "string") then return 0 end
local time = redis.call("TIME")
local now = tonumber(time[1]) * 1000000 + tonumber(time[2])
if global_type == "string" or client_type == "string" then
 if global_type == "string" then reserve_legacy(KEYS[1], tonumber(ARGV[1]), now, ARGV[4]) end
 if client_type == "string" then reserve_legacy(KEYS[2], tonumber(ARGV[2]), now, ARGV[4]) end
 return 0
end
local cutoff = now - tonumber(ARGV[3])
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", cutoff)
redis.call("ZREMRANGEBYSCORE", KEYS[2], "-inf", cutoff)
if redis.call("ZCARD", KEYS[1]) >= tonumber(ARGV[1]) or
   redis.call("ZCARD", KEYS[2]) >= tonumber(ARGV[2]) then return 0 end
redis.call("ZADD", KEYS[1], now, ARGV[5])
redis.call("ZADD", KEYS[2], now, ARGV[5])
redis.call("PEXPIRE", KEYS[1], ARGV[4])
redis.call("PEXPIRE", KEYS[2], ARGV[4])
return 1
`)

var providerAdmissionScript = redis.NewScript(legacyReservationScript + `
local used_type = redis.call("TYPE", KEYS[1]).ok
if used_type ~= "none" and used_type ~= "zset" and used_type ~= "string" then return 0 end
local time = redis.call("TIME")
local now = tonumber(time[1]) * 1000000 + tonumber(time[2])
if used_type == "string" then
 reserve_legacy(KEYS[1], tonumber(ARGV[1]), now, ARGV[3])
 return 0
end
local cutoff = now - tonumber(ARGV[2])
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", cutoff)
if redis.call("ZCARD", KEYS[1]) >= tonumber(ARGV[1]) then return 0 end
local remote = redis.call("GET", KEYS[2])
if remote and tonumber(remote) <= 0 then return 0 end
redis.call("ZADD", KEYS[1], now, ARGV[4])
redis.call("PEXPIRE", KEYS[1], ARGV[3])
if remote then redis.call("DECR", KEYS[2]) end
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
	if strings.TrimSpace(clientIdentity) == "" || clientLimit <= 0 || globalLimit <= 0 || window < time.Microsecond {
		return false, errors.New("shared endpoint budget configuration is invalid")
	}
	member, err := admissionMember()
	if err != nil {
		return false, fmt.Errorf("%w: admission identity: %v", ErrUnavailable, err)
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
		windowMicroseconds(window),
		windowTTL(window),
		member,
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
	if limit <= 0 || window < time.Microsecond {
		return false, errors.New("shared provider budget configuration is invalid")
	}
	member, err := admissionMember()
	if err != nil {
		return false, fmt.Errorf("%w: admission identity: %v", ErrUnavailable, err)
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
		windowMicroseconds(window),
		windowTTL(window),
		member,
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

func admissionMember() (string, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(nonce[:]), nil
}

func windowTTL(window time.Duration) int64 {
	ttl := window.Milliseconds()
	if window%time.Millisecond != 0 {
		ttl++
	}
	return ttl
}

func windowMicroseconds(window time.Duration) int64 {
	value := window.Microseconds()
	if window%time.Microsecond != 0 {
		value++
	}
	return value
}
