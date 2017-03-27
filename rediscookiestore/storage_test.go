package rediscookiestore

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/rmdashrf/go-misc/cookiejar2"
)

var (
	testCookie1 = &http.Cookie{
		Name:  "testCookie1",
		Value: "testCookie1 value",
	}

	testCookie2 = &http.Cookie{
		Name:  "testCookie2",
		Value: "testCookie2.value",
	}

	foobarUrl, _  = url.Parse("http://foobar.com")
	anotherUrl, _ = url.Parse("http://another.com")
)

func TestNoStorage(t *testing.T) {
	cj := cookiejar2.New(nil)
	cj.SetCookies(foobarUrl, []*http.Cookie{testCookie1})
	cj.SaveCookies()
	cookies := cj.Cookies(foobarUrl)
	if len(cookies) != 1 {
		t.Fatal("wrong number of cookies")
	}

	if cookies[0].Value != "testCookie1 value" {
		t.Fatal("wrong cookie")
	}

	entries := cj.Entries()
	if _, exists := entries["foobar.com"]["foobar.com;/;testCookie1"]; !exists {
		t.Fatal("Dont have expected entry")
	}
}

func TestRedisPersistence(t *testing.T) {
	tmpname := fmt.Sprintf("testRedisStore-%d", rand.Int())
	cl := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	redisStore := NewRedisCookieStore(cl, tmpname)

	cj := cookiejar2.New(&cookiejar2.Options{
		Storage: redisStore,
	})

	cj.SetCookies(foobarUrl, []*http.Cookie{testCookie1})
	cj.SaveCookies()

	cookies := cj.Cookies(foobarUrl)
	if len(cookies) != 1 {
		t.Fatal("wrong number of cookies")
	}

	if cookies[0].Value != "testCookie1 value" {
		t.Fatal("wrong cookie")
	}

	// Client 2
	redisStore2 := NewRedisCookieStore(cl, tmpname)
	cj2 := cookiejar2.New(&cookiejar2.Options{
		Storage: redisStore2,
	})

	cj2.SetCookies(anotherUrl, []*http.Cookie{testCookie2})
	cj2.SaveCookies()
	time.Sleep(500 * time.Millisecond)

	entries := cj.Entries()
	if len(entries) != 2 {
		t.Fatal("wrong number of entries")
	}

	if _, exists := entries["foobar.com"]["foobar.com;/;testCookie1"]; !exists {
		t.Fatal("Dont have expected entry")
	}
	if _, exists := entries["another.com"]["another.com;/;testCookie2"]; !exists {
		t.Fatal("Dont have expected entry")
	}

}
