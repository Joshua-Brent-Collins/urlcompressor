package storage

//_ here allows import for side effects/registration only
import (
    "github.com/go-redis/redis"
    _ "github.com/lib/pq"
    "database/sql"
    "fmt"
)

type StorageService struct {

    StorageMode int
    RClient *redis.Client
    DBClient *sql.DB
}

func (sService *StorageService) Store(hash string, url string) {
    sService.RClient.Set(hash, url, 0)
    qr, err := sService.DBClient.Query(`INSERT INTO short_urls(hash, url) VALUES($1, $2)`, hash, url)
    fmt.Println(qr)
    fmt.Println("-------")
    fmt.Println(err)
}

func (sService *StorageService) Lookup(hash string) string {
    cachedValue, _ := sService.RClient.Get(hash).Result()
    return cachedValue
}

func NewStorageService(connections map[string]string) *StorageService {
    redisConfig := &redis.Options {
        Addr: "localhost:6379",
        Password : "",
        DB : 0, //Default
    }
    dbConnection, tmp := sql.Open("postgres", connections["pgsql"])
    fmt.Println(tmp)
    return &StorageService{StorageMode: 0, DBClient: dbConnection, RClient: redis.NewClient(redisConfig)}
}
