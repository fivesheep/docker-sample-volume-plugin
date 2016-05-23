package main

import (
	i "golang.org/x/exp/inotify"
	r "gopkg.in/redis.v3"
	"testing"
)

var (
	testDriver *testRedisDriver
	testCache  map[uint32]string
	testClient *r.Client
)

type testRedisDriver struct{}

func (self *testRedisDriver) redisSet(client *r.Client, key, value string) error {
	return nil
}

func (self *testRedisDriver) redisGet(client *r.Client, key string) (string, error) {
	return "", nil
}

func (self *testRedisDriver) redisDel(client *r.Client, key string) error {
	return nil
}

func (self *testRedisDriver) readFile(filePath string) ([]byte, error) {
	testData := "test"
	return []byte(testData), nil
}

func init() {
	testDriver = &testRedisDriver{}
	testCache = make(map[uint32]string)
	testClient = nil
}

func TestCreateFileEvent(t *testing.T) {
	testEvent := i.Event{
		Mask:   i.IN_CREATE,
		Cookie: 0,
		Name:   "xyz",
	}
	err := handleEvent(&testEvent, testCache, testClient, testDriver)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteFileEvent(t *testing.T) {
	testEvent := i.Event{
		Mask:   i.IN_DELETE,
		Cookie: 0,
		Name:   "xyz",
	}
	err := handleEvent(&testEvent, testCache, testClient, testDriver)
	if err != nil {
		t.Error(err)
	}
}

func TestModifyFileEvent(t *testing.T) {
	testEvent := i.Event{
		Mask:   i.IN_MODIFY,
		Cookie: 0,
		Name:   "xyz",
	}
	err := handleEvent(&testEvent, testCache, testClient, testDriver)
	if err != nil {
		t.Error(err)
	}
}

func TestMoveFileEvent(t *testing.T) {
	testEvent := i.Event{
		Mask:   i.IN_MOVED_FROM,
		Cookie: 1234,
		Name:   "xyz",
	}
	err := handleEvent(&testEvent, testCache, testClient, testDriver)
	if err != nil {
		t.Error(err)
	}

	testEvent = i.Event{
		Mask:   i.IN_MOVED_TO,
		Cookie: 1234,
		Name:   "xyz",
	}
	err = handleEvent(&testEvent, testCache, testClient, testDriver)
	if err != nil {
		t.Error(err)
	}
}
