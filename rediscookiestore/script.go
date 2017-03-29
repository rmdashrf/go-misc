package rediscookiestore

import "github.com/go-redis/redis"

var (
	// SETANDPUB <setkey> <pubkey> <val> <token>
	// will set <setkey> to <val>, and then publish <token>
	// to the pubsub key <pubkey>
	scriptSetAndPublishSrc = `
local val = ARGV[1]
local token = ARGV[2]
local key = KEYS[1]
local publish_key = KEYS[2]

redis.call("SET", key, val)
redis.call("PUBLISH", publish_key, token)
return "OK"
`
	scriptSetAndPub = redis.NewScript(scriptSetAndPublishSrc)
)
