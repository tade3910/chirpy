package db

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Database struct {
	Chirps map[int]Chirp
	Users  map[int]User
}

type User struct {
	Id    int
	Email string
}
type Chirp struct {
	Id   int
	Body string
}

type Db struct {
	mu         sync.Mutex
	path       string
	nextId     int
	nextUserId int
}

func (database *Db) getDatabase() (*Database, bool) {
	database.mu.Lock()
	defer database.mu.Unlock()
	fileContent, err := os.ReadFile(database.path)
	if err != nil {
		fmt.Println("Problem reading file")
		return nil, false
	}
	currentDatabase := &Database{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
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

func (databse *Db) GetOne(chripId int) (Chirp, bool) {
	readDatabase, success := databse.getDatabase()
	if !success {
		return Chirp{}, false
	}
	chirp, ok := readDatabase.Chirps[chripId]
	if !ok {
		return Chirp{}, false
	}
	return chirp, true
}

func (db *Db) GetFormatedDatabase() ([]Chirp, bool) {
	database, success := db.getDatabase()
	if !success {
		return nil, false
	}
	formatedDatabse := make([]Chirp, len(database.Chirps))
	for key, chirp := range database.Chirps {
		formatedDatabse[key] = chirp
	}
	return formatedDatabse, true
}

func (db *Db) UpdateChirps(data string) (Chirp, bool) {
	database, success := db.getDatabase()
	if !success {
		fmt.Println("Problem getting database")
		return Chirp{}, false
	}
	db.mu.Lock()
	nextChirp := Chirp{
		Id:   db.nextId,
		Body: data,
	}
	database.Chirps[db.nextId] = nextChirp
	db.mu.Unlock()
	success = db.updateDatabase(database)
	if !success {
		fmt.Println("Problem getting database")
		return Chirp{}, false
	}
	db.mu.Lock()
	db.nextId++
	db.mu.Unlock()
	return nextChirp, true
}

func (db *Db) UpdateUsers(email string) (User, bool) {
	database, success := db.getDatabase()
	if !success {
		fmt.Println("Problem getting database")
		return User{}, false
	}
	db.mu.Lock()
	nextUser := User{
		Id:    db.nextUserId,
		Email: email,
	}
	database.Users[db.nextUserId] = nextUser
	db.mu.Unlock()
	success = db.updateDatabase(database)
	if !success {
		fmt.Println("Problem getting database")
		return User{}, false
	}
	db.mu.Lock()
	db.nextUserId++
	db.mu.Unlock()
	return nextUser, true
}

func (db *Db) updateDatabase(database *Database) bool {
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
