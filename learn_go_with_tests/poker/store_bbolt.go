package poker

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

// NewPlayerStoreBolt opens, and creates if necessary, the given BoltDB file.
func NewPlayerStoreBolt(path string) (*PlayerStoreBolt, error) {
	log.Println("Open BoltDB database file:", path)
	db, err := setupBoltDB(path)
	if err != nil {
		return nil, err
	}
	return &PlayerStoreBolt{db}, nil
}

// League implements PlayerStore.League
func (s *PlayerStoreBolt) League() (League, error) {
	var league League
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.ForEach(func(name, value []byte) error {
			score, err := strconv.Atoi(string(value))
			if err != nil {
				return fmt.Errorf("converting score: %w", err)
			}
			league = append(league, Player{string(name), score})
			return nil
		})
		return err
	})
	return league, err
}

// Score implements PlayerStore.Score
func (s *PlayerStoreBolt) Score(name string) (int, error) {
	var score int
	var err error
	err = s.db.View(func(tx *bolt.Tx) error {
		score, err = s.getScore(tx, name)
		return nil
	})
	return score, err
}

// SetScore implements PlayerStore.SetScore
func (s *PlayerStoreBolt) SetScore(name string, score int) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		return s.setScore(tx, name, score)
	})
	return err
}

// RecordWin implements PlayerStore.RecordWin
func (s *PlayerStoreBolt) RecordWin(name string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		// Fetch current
		score, err := s.getScore(tx, name)
		if err != nil {
			return err
		}

		// Increment and save
		score++
		return s.setScore(tx, name, score)
	})
	return err
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

// setScore is broken out so that it can be called from within a transaction
func (s *PlayerStoreBolt) setScore(tx *bolt.Tx, name string, score int) error {
	value := strconv.Itoa(score)
	err := tx.Bucket([]byte(bucketName)).Put([]byte(name), []byte(value))
	if err != nil {
		return fmt.Errorf("updating score: %v", err)
	}
	return nil
}

// setupBoltDB opens database (creating if necessary) and ensures scores bucket exists
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
