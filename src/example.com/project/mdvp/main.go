// main package contains the implementation of my docker volume plugin, a.k.a.
// mdvp
package main

import (
	"fmt"
	v "github.com/docker/go-plugins-helpers/volume"
	"log"
	"os"
)

var (
	pluginName = "mdvp"
)

// volumeInfo keeps a track of whetehr the volume is mounted or not
type volumeInfo struct {
	isMounted bool
}

// MyDockerVolumePlugin implements the volume Driver interface
type MyDockerVolumePlugin struct {
	// the root path under which all the volumes are stored by this driver
	rootPathOnDisk string
	// the map of the volume name to its VolumeInfo
	metadata map[string]*volumeInfo
}

// --- Plugin handlers start here ---

func (self *MyDockerVolumePlugin) Create(req v.Request) v.Response {
	log.Printf("Create: %+v\n", req)
	resp := v.Response{}
	if req.Name == "" {
		resp.Err = "Volume name not specified!"
		return resp
	}

	// check if the entry already exists!
	info, ok := self.metadata[req.Name]
	if ok {
		// pre-created volume; noop
		// "or should I return an error as a volume creation request with
		// same volume-name has come up twice now. This relates to when does
		// docker call create event: during docker run or only once when it
		// encounters a new volume. In short, does docker keep track of volume
        // names?" - docker caches! :)
		log.Printf("Create called for pre-existing volume, info: %+v\n", info)
	} else {
		// new volume create req
		self.metadata[req.Name] = &volumeInfo{
			isMounted: false,
		}
		// create the directory for this volume
		dirPath := fmt.Sprintf("%s/%s", self.rootPathOnDisk, req.Name)
		// TODO - hardcoded permission bits
		err := os.Mkdir(dirPath, 0777)
		if err != nil {
			if errr, ok := err.(*os.PathError); ok {
				log.Println("Path Error while calling mkdir")
				log.Println("Operation: ", errr.Op)
				log.Println("Path: ", errr.Path)
				log.Println("Err: ", errr.Err)
			} else {
				// some other error
				log.Println(err)
			}
			resp.Err = "Cannot create volume!"
		}
	}
	return resp
}

func (self *MyDockerVolumePlugin) Remove(req v.Request) v.Response {
	log.Printf("Remove: %+v\n", req)
	resp := v.Response{}
	if req.Name == "" {
		resp.Err = "Volume name not specified!"
		return resp
	}

	info, ok := self.metadata[req.Name]
	if ok {
		if info.isMounted {
			// send an error as volume is still in use
			resp.Err = "Remove failed as volume is still in use!"
		} else {
			// delete from cache
			delete(self.metadata, req.Name)
			// delete the dir
			dirPath := fmt.Sprintf("%s/%s", self.rootPathOnDisk, req.Name)
			err := os.RemoveAll(dirPath)
			if err != nil {
				if errr, ok := err.(*os.PathError); ok {
					log.Println("Path Error while calling rmdir")
					log.Println("Operation: ", errr.Op)
					log.Println("Path: ", errr.Path)
					log.Println("Err: ", errr.Err)
				} else {
					// some other error
					log.Println(err)
				}
				resp.Err = "Cannot remove volume!"
			}
		}
	} else {
		resp.Err = "Removing a volume that was never created!"
	}
	return resp
}

func (self *MyDockerVolumePlugin) Mount(req v.Request) v.Response {
	log.Printf("Mount: %+v\n", req)
	resp := v.Response{}
	if req.Name == "" {
		resp.Err = "Volume name not specified!"
		return resp
	}

	// check if this volume was created before
	info, ok := self.metadata[req.Name]
	if ok {
		// is it already mounted and in use?
		if info.isMounted {
			// return an error
			resp.Err = "Mount operation on an already mounted volume!"
		} else {
			info.isMounted = true
			// populate the mount point field
			resp.Mountpoint = fmt.Sprintf("%s/%s", self.rootPathOnDisk, req.Name)
		}
	} else {
		resp.Err = "Mount operation in an unknown volume!"
	}
	return resp
}

func (self *MyDockerVolumePlugin) Path(req v.Request) v.Response {
	log.Printf("Path: %+v\n", req)
	resp := v.Response{}
	if req.Name == "" {
		resp.Err = "Volume name not specified!"
		return resp
	}

	// check if this volume was created
	_, ok := self.metadata[req.Name]
	if ok {
		resp.Mountpoint = fmt.Sprintf("%s/%s", self.rootPathOnDisk, req.Name)
	} else {
		resp.Err = "Path query on an unknown volume!"
	}
	return resp
}

func (self *MyDockerVolumePlugin) Unmount(req v.Request) v.Response {
	log.Printf("Unmount: %+v\n", req)
	resp := v.Response{}
	if req.Name == "" {
		resp.Err = "Volume name not specified!"
		return resp
	}

	// check if the volume exists and was previously mounted
	info, ok := self.metadata[req.Name]
	if ok {
		if info.isMounted {
			info.isMounted = false
		} else {
			resp.Err = "Unmounting a volume that was never mounted!"
		}
	} else {
		resp.Err = "Unmounting an unknown volume!"
	}
	return resp
}

func (self *MyDockerVolumePlugin) Get(req v.Request) v.Response {
	log.Printf("Get: %+v\n", req)
	resp := v.Response{}
	if req.Name == "" {
		resp.Err = "Volume name not specified!"
		return resp
	}

	// check if the volume exists and was previously mounted
	_, ok := self.metadata[req.Name]
	if ok {
		// if info.isMounted { - Commenting this as we need to send this info
		// to docker daemon even after unmount, for eg. when running a docker
		// volume rm command
		resp.Volume = &v.Volume{
			Name:       req.Name,
			Mountpoint: fmt.Sprintf("%s/%s", self.rootPathOnDisk, req.Name),
		}
		//}
	} else {
		resp.Err = "Cannot get volume info of an unknown volume!"
	}
	return resp
}

func (self *MyDockerVolumePlugin) List(req v.Request) v.Response {
	log.Printf("List: %+v\n", req)
	resp := v.Response{}
	var volume v.Volume
	for key := range self.metadata {
		volume = v.Volume{
			Name:       key,
			Mountpoint: fmt.Sprintf("%s/%s", self.rootPathOnDisk, key),
		}
		resp.Volumes = append(resp.Volumes, &volume)
	}
	return resp
}

// --- Plugin handlers end here ---

// NewMyDockerVolumePlugin returns the volume driver object
func NewMyDockerVolumePlugin(path string) (*MyDockerVolumePlugin, error) {
	mdnp := &MyDockerVolumePlugin{
		rootPathOnDisk: path,
		metadata:       make(map[string]*volumeInfo),
	}
	return mdnp, nil
}

func main() {
	driver, err := NewMyDockerVolumePlugin("/tmp")
	if err != nil {
		log.Fatalf("ERROR: %s init failed!", pluginName)
	}
	requestHandler := v.NewHandler(driver)
	requestHandler.ServeTCP(pluginName, ":3004")
}
