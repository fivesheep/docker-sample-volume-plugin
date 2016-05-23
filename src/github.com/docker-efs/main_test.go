package main

import (
	"fmt"
	v "github.com/docker/go-plugins-helpers/volume"
	"log"
	"os"
	"testing"
)

var (
	testDriver *MyDockerVolumePlugin
	err        error
	root       string
	volumeName string
)

func init() {
	root = "/tmp"
	volumeName = "test-driver"
	testDriver, err = NewMyDockerVolumePlugin(root)
	if err != nil {
		log.Fatalf("Couldn't init mdvp volume driver. Exiting!")
	}
}

// createVolumeTest verifies that the folder has been created under root dir to
// act as volume for container
func createVolumeTest(volumeName string) bool {
	// driver should have created the volume folder under root
	_, err := os.Stat(fmt.Sprintf("%s/%s", root, volumeName))
	if err == nil {
		return true
	} else {
		log.Printf("createVolumeTest: %s\n", err)
	}
	return false
}

func removeVolumeTest(volumeName string) bool {
	// driver should have deleted the volume folder under root
	_, err := os.Stat(fmt.Sprintf("%s/%s", root, volumeName))
	if err != nil {
		if os.IsNotExist(err) {
			return true
		} else {
			// some other OS error
			log.Printf("removeVolumeTest: %s\n", err)
		}
	} else {
		log.Printf("remove volume operation seems to have failed!")
	}
	return false
}

func TestMain(m *testing.M) {
	// create a volume
	createReq := v.Request{
		Name:    volumeName,
		Options: make(map[string]string),
	}
	createResp := testDriver.Create(createReq)
	if createResp.Err == "" {
		createVolumeTest(volumeName)
	} else {
		log.Fatalf("create volume failed: %s\n", createResp.Err)
	}

	// run the tests
	exitCode := m.Run()

	// remove a volume
	removeReq := v.Request{
		Name: volumeName,
	}
	removeResp := testDriver.Remove(removeReq)
	if removeResp.Err == "" {
		removeVolumeTest(volumeName)
	} else {
		log.Fatalf("remove volume failed: %s\n", removeResp.Err)
	}

	// exit with the exitCode
	os.Exit(exitCode)
}

func TestMount(t *testing.T) {
	mountReq := v.Request{
		Name: volumeName,
	}
	mountResp := testDriver.Mount(mountReq)
	if mountResp.Err != "" {
		t.Errorf("Mount got an error: %s\n", mountResp.Err)
		removeVolumeTest(volumeName)
	} else {
		// check for the mountpoint
		expectedMountpoint := fmt.Sprintf("%s/%s", root, volumeName)
		if mountResp.Mountpoint != expectedMountpoint {
			t.Errorf("Mount got a wrong mountpoint: %s\n", mountResp.Mountpoint)
		}
	}
}

func TestUnmount(t *testing.T) {
	unmountReq := v.Request{
		Name: volumeName,
	}
	unmountResp := testDriver.Unmount(unmountReq)
	if unmountResp.Err != "" {
		t.Errorf("Unmount got an err: %s\n", unmountResp.Err)
	}
}

/*
TODO - how do we run tests sequentially? What if these run after Unmount is called?

func TestList(t *testing.t) {}
func TestGet(t *testing.t) {}
func TestPath(t *testing.t) {}
*/
