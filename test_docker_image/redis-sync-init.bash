#!/bin/bash

SUPERVISOR_CONF_FILE=/etc/supervisor/supervisord.conf
cat <<EOF > $SUPERVISOR_CONF_FILE
; supervisor config file

[unix_http_server]
file=/var/run/supervisor.sock
chmod=0700

[supervisord]
logfile=/var/log/supervisor/supervisord.log
logfile_maxbytes=50MB
logfile_backups=2
loglevel=debug
pidfile=/var/run/supervisord.pid
umask=022
nodaemon=true

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisorctl]
serverurl=unix:///var/run/supervisor.sock

[program:file-sync-redis]
command=/usr/local/bin/inotify

[program:redis-server]
command=redis-server

EOF

# start our service
/usr/bin/supervisord

