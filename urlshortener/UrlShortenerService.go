package urlshortener

import (
    "github.com/go-redis/redis"
    "crypto/sha256"
    "encoding/base64"
    "strings"
)

type UrlShortener struct {

    UrlIdLength int64
    RClient *redis.Client
    BaseUrl string
}

func (us *UrlShortener)SetBaseUrl(baseUrl string) {
    if strings.LastIndex(baseUrl, "/") != (len(baseUrl) - 1) {
        us.BaseUrl = baseUrl + "/"
    }
}

func (us *UrlShortener) GenerateHash(input string, rounds int) string {
    hash := sha256.New()
    hash.Write([]byte(input))

    input = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash.Sum(nil))
    if rounds != 1 {
        input = us.GenerateHash(input, rounds - 1 )
    }
    input = strings.Replace(strings.Replace(input, "_", "", -1), "-", "", -1)
    return input[0:(us.UrlIdLength)]
}

func (us *UrlShortener) ShortenUrl(urlToShorten string) string {
    urlIdHash := us.GenerateHash(urlToShorten, 2)
    us.RClient.Set(urlIdHash, urlToShorten, 0)
    return us.BaseUrl + urlIdHash
}

func (us *UrlShortener) GetOriginalUrl(urlIdHash string) string {
    cachedValue, _ := us.RClient.Get(urlIdHash).Result()
    return cachedValue
}