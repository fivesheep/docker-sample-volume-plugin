# docker build -t krish/test:v1 .
# docker run -d -v redis-vol:/foo --volume-driver=mdvp --name test-redis1 --rm=true redis

# TODO(Krish): move to alpine once testing is complete

FROM redis
RUN apt-get update \
  && apt-get -y install supervisor inotify-tools --no-install-recommends \
  && rm -rf /etc/supervisor/conf.d \
  && rm -rf /etc/supervisor/supervisord.conf \
  && mkdir /foo
# Copy dependencies
COPY bin/inotify /usr/local/bin/
COPY redis-sync-init.bash /usr/local/bin/

# Clean to reduce image size
RUN apt-get -y autoremove \
    && apt-get -y autoclean \
    && apt-get clean
ENTRYPOINT ["/usr/local/bin/redis-sync-init.bash"]

