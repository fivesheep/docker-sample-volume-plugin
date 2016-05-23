package main

import (
	"errors"
	i "golang.org/x/exp/inotify"
	r "gopkg.in/redis.v3"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var (
	ErrSetFailed = errors.New("Error during SET")
	ErrGetFailed = errors.New("Error during GET")
	ErrDelFailed = errors.New("Error during DEL")
	ErrFileRead  = errors.New("Error during file read")
)

type RedisDriver interface {
	redisSet(*r.Client, string, string) error
	redisGet(*r.Client, string) (string, error)
	redisDel(*r.Client, string) error
	readFile(string) ([]byte, error)
}

type redisDriver struct{}

// redisSet is called to set values in redis. The filepath & value are
// parameters to this function, alongwith the redis client
func (self *redisDriver) redisSet(client *r.Client, key, value string) error {
	// expiration will always be zero
	expiration := 0 * time.Second
	status := client.Set(getKeyName(key), value, expiration)
	if status.Err() != nil {
		log.Printf("Error during SET: %s\n", status.Err())
		return status.Err()
	}
	return nil
}

// redisGet is called to get values from redis. The filepath is passed as key
// to the function, alongwith the redis client
func (self *redisDriver) redisGet(client *r.Client, key string) (string, error) {
	stringCmd := client.Get(getKeyName(key))
	value, err := stringCmd.Result()
	if err != nil {
		log.Printf("Error during GET: %s\n", err)
		return "", err
	}
	if err == r.Nil {
		log.Printf("Key %s does not exist\n", key)
		return "", nil
	}
	return value, nil
}

// redisDel is called to delete values from redis. The filepath is passed as
// key to the function, alongwith the redis client
func (self *redisDriver) redisDel(client *r.Client, key string) error {
	intCmd := client.Del(getKeyName(key))
	if intCmd.Err() != nil {
		log.Printf("Error during DEL: %s\n", intCmd.Err())
		return intCmd.Err()
	}
	return nil
}

func (self *redisDriver) readFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

// getKeyName splits the file path passed as parameter to derive the file name,
// which is returned
func getKeyName(filePath string) string {
	absPath := strings.Split(filePath, "/")
	fileName := absPath[len(absPath)-1]
	log.Printf("Keyname: %s\n", fileName)
	return fileName
}

// handleEvent sifts the inotify events and only handles the interesting ones.
// A move event is a corner case, where is the cookie is used to relate 2
// corresponding move_from and move_to events.
func handleEvent(event *i.Event, cache map[uint32]string, client *r.Client, driver RedisDriver) error {
	switch event.Mask {
	case i.IN_CREATE:
		log.Printf("Create: %s\n", event.Name)
		if driver.redisSet(client, event.Name, "") != nil {
			return ErrSetFailed
		}
	case i.IN_DELETE:
		log.Printf("Delete: %s\n", event.Name)
		if driver.redisDel(client, event.Name) != nil {
			return ErrDelFailed
		}
	case i.IN_MODIFY:
		log.Printf("Modify: %s\n", event.Name)
		// read the file data
		value, err := driver.readFile(event.Name)
		if err != nil {
			return ErrFileRead
		}
		if driver.redisSet(client, event.Name, string(value[:len(value)-1])) != nil {
			return ErrSetFailed
		}
	case i.IN_MOVED_FROM:
		log.Printf("Move src: %s, Cookie: %d\n", event.Name, event.Cookie)
		value, err := driver.redisGet(client, event.Name)
		if err != nil {
			return ErrGetFailed
		}
		cache[event.Cookie] = value
		if driver.redisDel(client, event.Name) != nil {
			return ErrDelFailed
		}
	case i.IN_MOVED_TO:
		log.Printf("Move dst: %s, Cookie: %d\n", event.Name, event.Cookie)
		value := cache[event.Cookie]
		delete(cache, event.Cookie)
		if driver.redisSet(client, event.Name, string(value)) != nil {
			return ErrSetFailed
		}
		//default: - results in a lot of spurious prints
		//	log.Printf("Unhandled event: %s\n", event)
	}
	return nil
}

// main sets up a watch on /foo inside the container, conencts to the redis
// instance, sets up a cache for handling inter-linked inotify events, &
// watches the /foo directory forever
func main() {
	watchDir := "/foo"
	cache := make(map[uint32]string)
	client := r.NewClient(&r.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// create our redis driver instance
	driver := &redisDriver{}

	watcher, err := i.NewWatcher()
	if err != nil {
		log.Fatal("Could not setup watcher: ", err)
	}
	err = watcher.Watch(watchDir)
	if err != nil {
		log.Fatal("Could not setup watch: ", err)
	}
	for {
		select {
		case ev := <-watcher.Event:
			if err := handleEvent(ev, cache, client, driver); err != nil {
				log.Fatal(err)
			}
		case err := <-watcher.Error:
			log.Fatal("Error from watcher: ", err)
		}
	}
}
