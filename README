# mdvp

-------------------------------------------
:Testing of my docker volume plugin (mdvp):
-------------------------------------------
I use the "/tmp" folder as the root volume. Every subsequent volume is created under /tmp.
Copy mdvp.json to /etc/docker/plugins/
Start the plugin: sudo ./bin/main 
Restart docker service.
Container 1:
  docker run -v redis-db:/data --volume-driver=mdvp --name test-redis1 --rm=true redis
  docker exec -it test-redis1 bash
  redis-cli
  SET mykey "Hello"
  SAVE
  exit the cli and stop this container
Container 2:
  docker run -v redis-db:/data --volume-driver=mdvp --name test-redis2 redis
  docker exec -it test-redis2 bash
  redis-cli
  GET mykey
  APPEND mykey " world" - optional
  SAVE - optional
  do NOT stop this container, start container 3
Container 3:
  docker run -v redis-db:/data --volume-driver=mdvp --name test-redis3 redis
  You will get an error:
    "docker: Error response from daemon: VolumeDriver.Mount: Mount operation on an already mounted volume!. See 'docker run --help'."
---------------------------------------------
:To delete the volumes at the end of testing:
---------------------------------------------
docker volume  rm redis-db
------------
:References:
------------
  https://github.com/docker/docker/blob/master/docs/extend/plugin_api.md
  https://github.com/docker/libnetwork/blob/master/docs/design.md
  https://github.com/docker/libnetwork/blob/master/docs/remote.md


