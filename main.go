package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

type Note struct {
	Content   string
	CreatedAt time.Time
	Author    string
}

//making a login like system with -b flag
//login will be used as users bucket
var (
	Bucket string
)

func init() {
	flag.StringVar(&Bucket, "b", "", "Bucket name")
	flag.Parse()
}

func main() {
	var choice byte
	var note Note
	fmt.Printf("Menu:\n\t1. Create and store new note\n\t2. Show note\n\t3. Notes list\n\t0. Exit\n")

	reader := bufio.NewReader(os.Stdin)

	db, err := bolt.Open("my.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//initialize bucket if it's a first user's entry
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(Bucket))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	for {
		fmt.Printf(">>> ")
		fmt.Scanf("%d", &choice)
		switch choice {
		case 1:
			fmt.Printf("Note title: ")
			title, _ := reader.ReadString('\n')
			fmt.Printf("Note content: ")
			note.Content, _ = reader.ReadString('\n')
			fmt.Printf("Note written by: ")
			note.Author, _ = reader.ReadString('\n')
			note.CreatedAt = time.Now()

			encoded, err := encodeNote(note)
			if err != nil {
				log.Fatal(err)
			}
			err = makeEntry(db, []byte(title), encoded)
			if err != nil {
				log.Fatal(err)
			}

		case 2:
			fmt.Printf("Enter note's title: ")
			title, _ := reader.ReadString('\n')
			val, err := getNote(db, []byte(title))
			if err != nil {
				log.Fatal(err)
			}

			decodedStruct, err := decodeNote(val)
			if err != nil {
				log.Fatal(err)
			}
			title = strings.TrimSuffix(title, "\n")
			decodedStruct.Content = strings.TrimSuffix(decodedStruct.Content, "\n")
			decodedStruct.Author = strings.TrimSuffix(decodedStruct.Author, "\n")

			fmt.Printf("Title: %s\nNote: %s\nWritten by %s at %v\n", title, decodedStruct.Content, decodedStruct.Author, decodedStruct.CreatedAt)

		case 3:
			err := getKeys(db)
			if err != nil {
				log.Fatal(err)
			}

		case 0:
			os.Exit(1)
		default:
			fmt.Println("Invalid input")
		}
	}
}

func encodeNote(n Note) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(n); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func decodeNote(b []byte) (Note, error) {
	var out Note
	r := bytes.NewReader(b)
	dec := gob.NewDecoder(r)
	if err := dec.Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

func makeEntry(db *bolt.DB, key []byte, value []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(Bucket))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", Bucket)
		}

		err := bucket.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func getNote(db *bolt.DB, key []byte) ([]byte, error) {
	var val []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(Bucket))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", Bucket)
		}

		val = bucket.Get(key)
		return nil
	})
	if err != nil {
		return val, err
	}
	return val, nil
}

func getKeys(db *bolt.DB) error {
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(Bucket))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", Bucket)
		}
		c := bucket.Cursor()
		fmt.Printf("Available notes:\n")
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			k = []byte(strings.TrimSuffix(string(k), "\n"))
			fmt.Printf("\t%s\n", k)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
