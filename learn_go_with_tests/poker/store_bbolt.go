package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
)

const bucketName = "scores"

// PlayerStoreBolt uses bbolt to implement PlayerStore
type PlayerStoreBolt struct {
	db *bolt.DB
}

func NewPlayerStoreBolt(path string) *PlayerStoreBolt {
	log.Println("Open BoltDB database file:", path)
	db, err := setupBoltDB(path)
	if err != nil {
		log.Fatal(err)
	}
	return &PlayerStoreBolt{db}
}

// GetScore fetches the current score for the named player
func (s *PlayerStoreBolt) GetScore(name string) (int, error) {
	var score int
	var err error
	err = s.db.View(func(tx *bolt.Tx) error {
		score, err = s.getScore(tx, name)
		return nil
	})
	return score, err
}

// getScore is broken out so that it can be called from within RecordWin's transaction
func (s *PlayerStoreBolt) getScore(tx *bolt.Tx, name string) (int, error) {
	value := tx.Bucket([]byte(bucketName)).Get([]byte(name))
	if len(value) == 0 {
		return 0, nil
	}
	score, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, fmt.Errorf("converting score: %w", err)
	}
	return score, nil
}

// RecordWin increments the score of the named player
func (s *PlayerStoreBolt) RecordWin(name string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		// Fetch current
		score, err := s.getScore(tx, name)
		if err != nil {
			return err
		}
		score += 1

		// Increment and save
		value := strconv.Itoa(score)
		err = tx.Bucket([]byte(bucketName)).Put([]byte(name), []byte(value))
		if err != nil {
			return fmt.Errorf("updating score: %v", err)
		}
		return nil
	})
	return err
}

func setupBoltDB(path string) (*bolt.DB, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}
	db, err := bolt.Open(path, 0666, options)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("creating bucket %s: %w", bucketName, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	return db, nil
}
