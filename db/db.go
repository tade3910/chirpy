package db

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Database struct {
	Chirps     map[int]Chirp
	Users      map[string]*User
	IDUsersMap map[int]*User
	Sessions   map[string]Session
}

type Session struct {
	User    *User
	Expires time.Time
}

type User struct {
	Password []byte
	PlainUser
}

type PlainUser struct {
	Id            int
	Email         string
	Is_chirpy_red bool
}

type Chirp struct {
	Id       int
	Body     string
	AuthorId int
}

type Db struct {
	mu         sync.Mutex
	path       string
	nextId     int
	nextUserId int
}

func GetNewSession(user *User) Session {
	return Session{
		User:    user,
		Expires: time.Now().UTC().Add(time.Duration(60*24) * time.Hour),
	}
}

func (database *Db) GetNextId() int {
	database.mu.Lock()
	defer database.mu.Unlock()
	return database.nextId
}

func (database *Db) GetNextUserId() int {
	database.mu.Lock()
	defer database.mu.Unlock()
	return database.nextUserId
}

func (database *Db) addId() {
	database.mu.Lock()
	defer database.mu.Unlock()
	database.nextId++
}

func (database *Db) addUserId() {
	database.mu.Lock()
	defer database.mu.Unlock()
	database.nextUserId++
}

// Returns current database, handles empty json
func (database *Db) GetDatabase() (*Database, bool) {
	database.mu.Lock()
	defer database.mu.Unlock()
	fileContent, err := os.ReadFile(database.path)
	if err != nil {
		fmt.Println("Problem reading file")
		return nil, false
	}
	currentDatabase := &Database{
		Chirps:     map[int]Chirp{},
		Users:      map[string]*User{},
		IDUsersMap: map[int]*User{},
		Sessions:   map[string]Session{},
	}
	if len(fileContent) == 0 {
		return currentDatabase, true
	}
	err = json.Unmarshal(fileContent, currentDatabase)
	if err != nil {
		fmt.Println("Problem converting file bytes to databse struct")
		return nil, false
	}
	return currentDatabase, true
}

type databaseName string

const (
	ChirpDatabase databaseName = "chirp"
	UserDatabase  databaseName = "user"
	NoDatabase    databaseName = ""
)

func (db *Db) UpdateDatabase(database *Database, update databaseName) bool {
	ok := db.writeToJson(database)
	if ok {
		switch update {
		case ChirpDatabase:
			db.addId()
		case UserDatabase:
			db.addUserId()
		case NoDatabase:
			break
		default:
			return false
		}
	}
	return ok
}

func (db *Db) writeToJson(database *Database) bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	file, err := os.OpenFile(db.path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Problem opening file")
		return false
	}
	bytes, err := json.Marshal(database)
	if err != nil {
		fmt.Println("Problem converting database to bytes")
		return false
	}
	_, err = file.Write(bytes)
	if err != nil {
		fmt.Println("Problem writing to file")
		return false
	}
	return true
}

func GetDb() (*Db, bool) {
	path := "database.json"
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return nil, false
	}
	defer file.Close()
	newDb := &Db{
		path: "database.json",
	}
	return newDb, true
}
