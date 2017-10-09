package main

import (
    "github.com/go-redis/redis"
    "urlcompressor/urlshortener"
    "os"
    "strconv"
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    "io"
    "net/url"
    "strings"
 )

 var uShortener *urlshortener.UrlShortener

func main() {
    //UrlShortener setup
    redisConfig := &redis.Options {
        Addr: "localhost:6379",
        Password : "",
        DB : 0, //Default
    }
    urlIdLength, _ := strconv.ParseInt(os.Getenv("URL_ID_LENGTH"), 10, 64)
    baseUrl := os.Getenv("BASE_URL")
    redisClient := redis.NewClient(redisConfig)
    uShortener = &urlshortener.UrlShortener{UrlIdLength: urlIdLength, RClient: redisClient}
    uShortener.SetBaseUrl(baseUrl)

    server := SetupServer()
    RegisterHandlers()
    server.ListenAndServe()
}

//Structs for un/marshalling requests
type UrlReqRespEntity struct {
    UrlData string
}

func SetupServer() *http.Server {
    return &http.Server {
        Addr: ":12345",
    }
}

//Return 401 if auth fails http.StatusUnauthorized
func AuthenticateApiToken(originalHandler http.Handler) http.Handler {
    return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request){
        req.ParseForm()
        apiToken := req.Header.Get("API_KEY")
        if apiToken == "" {
            apiToken = req.Form.Get("apiKey")
        }
        if apiToken != os.Getenv("AUTH_TOKEN") {
            resp.WriteHeader(http.StatusUnauthorized)
            fmt.Fprint(resp, "Not Authorized!")
            return
        }
        originalHandler.ServeHTTP(resp, req)
    })
}

func HealthCheck(resp http.ResponseWriter, req *http.Request) {
    fmt.Fprint(resp, "Pong")
}

func HandleUrlRedirects(resp http.ResponseWriter, req *http.Request) {
    hashId := GetUrlHashPath(req.URL.String())
    redirectHandler := http.RedirectHandler(uShortener.GetOriginalUrl(hashId), http.StatusMovedPermanently)
    redirectHandler.ServeHTTP(resp, req)
}

func GenerateShortUrl(resp http.ResponseWriter, req *http.Request) {
    data := &UrlReqRespEntity{}
    json.Unmarshal(ParseBody(req.Body), data)
    data.UrlData = uShortener.ShortenUrl(data.UrlData)
    //This is going to be []byte so we will use Write() directly
    JSONResp, _ := json.Marshal(data)
    resp.Header().Set("Content-Type", "application/json")
    resp.Write(JSONResp)
}

func FindLongUrl(resp http.ResponseWriter, req *http.Request) {
    data := &UrlReqRespEntity{}
    json.Unmarshal(ParseBody(req.Body), data)
    tmp := GetUrlHashPath(data.UrlData)
    fmt.Fprint(resp, uShortener.GetOriginalUrl(tmp))
}

func RegisterHandlers() {
    http.Handle("/", http.HandlerFunc(HandleUrlRedirects))
    http.Handle("/Ping", http.HandlerFunc(HealthCheck))
    http.Handle("/GenerateShortUrl", AuthenticateApiToken(http.HandlerFunc(GenerateShortUrl)))
    http.Handle("/FindLongUrl", AuthenticateApiToken(http.HandlerFunc(FindLongUrl)))
}

func GetUrlHashPath(urlData string) string {
    parsedUrl, _ := url.Parse(urlData)
    return strings.Replace(parsedUrl.Path, "/", "", 1)
}

func ParseBody(strm io.Reader) []byte {
    dataBuffer := &bytes.Buffer{}
    dataBuffer.ReadFrom(strm)
    return dataBuffer.Bytes()
}
