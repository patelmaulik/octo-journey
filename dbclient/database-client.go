package dbclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"

	"patelmaulik.com/maulik/v1/models"

	bitcask "github.com/prologic/bitcask"
)

type DatabaseRawClient struct {
	database *bitcask.Bitcask
}

type IDatabaseRepository interface {
	CheckConection() bool
	OpenDatabase()
	QueryAccount(accountId string) (models.Account, error)
	SeedDatabase()
}

type DatabaseRepository struct {
	DbClient *DatabaseRawClient
}

func NewDbClient() *DatabaseRawClient {
	gc := &DatabaseRawClient{}
	return gc
}

func ensureConnectionExists(repo *DatabaseRepository) {
	if repo.DbClient == nil {
		repo.DbClient = NewDbClient()
	}
}

// CheckConection test connection
func (repo *DatabaseRepository) CheckConection() bool {
	return repo.DbClient.database.RLocked()
}

// SeedDatabase to start off with
func (repo *DatabaseRepository) SeedDatabase() {
	var total = 100
	for i := 0; i < total; i++ {
		key := strconv.Itoa(1000 + i)
		acc := models.Account{
			Id:   key,
			Name: "Person_" + strconv.Itoa(i),
		}

		jsonBytes, _ := json.Marshal(acc)
		repo.DbClient.database.Put([]byte(key), jsonBytes)
	}

	fmt.Printf("Seeded %v fake accounts\n", total)
}

// OpenDatabase using open database
func (repo *DatabaseRepository) OpenDatabase() {
	ensureConnectionExists(repo)

	var err error
	repo.DbClient.database, err = bitcask.Open("./accounts.db") // /tmp/
	if err != nil {
		log.Fatal(err)
	}
}

// QueryAccount using account-id
func (repo *DatabaseRepository) QueryAccount(accountId string) (models.Account, error) {
	account := models.Account{}

	var accountBytes []byte
	var err error

	if accountBytes, err = repo.DbClient.database.Get([]byte(accountId)); err != nil {
		return account, err
	}

	if accountBytes == nil {
		return account, fmt.Errorf("No account for %v", accountId)
	}

	json.Unmarshal(accountBytes, &account)

	account.ServedBy = GetIP()

	return account, nil
}

// GetIP - get ip address from loopback
func GetIP() string {
	_, err := net.InterfaceAddrs()
	if err != nil {
		return "error"
	}

	return "127.0.0.1"
	/*
		for _, address := range addr {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	*/

	// panic("Unable to determine loopback interface address IP")
}
