package urlshortener

//_ here allows import for side effects/registration only
import (
    "urlcompressor/storage"
    "crypto/sha256"
    "encoding/base64"
    "strings"
)

type UrlShortener struct {

    UrlIdLength int64
    StorageServiceClient *storage.StorageService
    BaseUrl string
}

func (us *UrlShortener) SetBaseUrl(baseUrl string) {
    if strings.LastIndex(baseUrl, "/") != (len(baseUrl) - 1) {
        us.BaseUrl = baseUrl + "/"
    } else {
        us.BaseUrl = baseUrl
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
    us.StorageServiceClient.Store(urlIdHash, urlToShorten)
    return us.BaseUrl + urlIdHash
}

func (us *UrlShortener) GetOriginalUrl(urlIdHash string) string {
    return us.StorageServiceClient.Lookup(urlIdHash)
}

func CreateNewShortener(connections map[string]string, urlHashLength int64, baseUrl string) *UrlShortener {

    urlShortenerPtr := &UrlShortener{UrlIdLength: urlHashLength, StorageServiceClient: storage.NewStorageService(connections)}
    urlShortenerPtr.SetBaseUrl(baseUrl)
    return urlShortenerPtr
}