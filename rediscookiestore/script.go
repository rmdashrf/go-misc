package rediscookiestore

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/rmdashrf/go-misc/cookiejar2"
)

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

func SetCookies(r *redis.Client, key string, entries cookiejar2.CookieEntries, id string) (err error) {
	var contents []byte
	contents, err = json.Marshal(entries)
	if err != nil {
		return
	}

	storeName := StoreName(key)
	invalidationName := InvalidationName(key)
	err = scriptSetAndPub.Run(r, []string{storeName, invalidationName}, contents, id).Err()
	return

}

func StoreName(key string) string {
	return fmt.Sprintf("%s:store", key)
}

func InvalidationName(key string) string {
	return fmt.Sprintf("%s:invalidation", key)
}
