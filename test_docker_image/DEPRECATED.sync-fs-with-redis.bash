#!/bin/bash

# DEPRECATED as inotifywait does not support cookie for 'move' functions

# Krish: 2016-05-02
# Prerequisite: sudo apt-get install inotify-tools
# Script that waits for events to occur in a mapped dir and updates redis
# corresponding to the filesystem changes.
# Assumption that there is no directory level action in the mapped volume;
# only files are created and destroyed and modified.
# File names are the keys and file data are the values for redis.

WATCH_DIR=/foo
DEBUG_MODE=1 # can be 0|1
# global map that keeps track of move operations
declare -A cache


function debug_log() {
    if [[ $# -ne 1 ]]; then
        # noop, nothing to log
        return
    fi
    if [[ $DEBUG_MODE -eq 1 ]]; then
        echo $1
    fi
}

function move_from() {
    if [[ $# -ne 2 ]]; then
        return
    fi
    file_path=$1
    file_name=$2
    value=`redis-cli get $file_name`
    # using the inode of the file
    inode=`stat -c "%i" $file_path/$file_name`
    ${cache[$inode]}=$value
    debug_log "deleting $file_name"
    redis-cli del $file_name
}

function move_to() {
    if [[ $# -ne 2 ]]; then
        return
    fi
    file_path=$1
    file_name=$2
    # get the inode
    inode=`stat -c "%i" $file_path/$file_name`
    # get the value from cache
    value=$cache[$inode]
    redis-cli set $file_name $value
    unset $cache[$inode] 
}

function print_cache() {
    for key in "${!cache[@]}"; do
        echo "$key - ${cache[$key]}"
    done
}

inotifywait --monitor \
  --recursive \
  --quiet \
  --format '%w %f %e' \
  --event modify \
  --event move \
  --event create \
  --event delete \
  $WATCH_DIR | while read file_path file_name event
do
    debug_log $file_path
    debug_log $file_name
    debug_log $event
    case $event in
        DELETE)
            # file deletion event; delete corresponding kv from redis
            debug_log "Deleting $file_name"
            redis-cli del $file_name
            ;;
        CREATE)
            # file creation event; create the corresponding kv in redis
            # default numm value is empty strings, as we are only working with
            # strings for now
            debug_log "Creating $file_name"
            redis-cli set $file_name ""
            ;;
        MODIFY)
            # file modification event; change the corresponding kv in redis
            value=`cat $file_path/$file_name`
            debug_log "DEBUG: Updating $file_name with $value"
            redis-cli set $file_name $value
            ;;
        MOVED_FROM)
            # file name changed event; delete the kv in redis, cache the value
            # from redis
            move_from $file_path $file_name
            print_cache
            ;;
        MOVED_TO)
            # file name changed event; create the k in redis, read from our
            # local cache to populate the value
            moved_to $file_path $file_name
            print_cache
            ;;
        *)
            debug_log "Unknown notification!" 
            ;;
    esac
done

exit 0

