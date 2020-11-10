FROM golang:1.15 as builder
COPY ./ /usr/src/cacheman
WORKDIR /usr/src/cacheman
RUN go build

FROM golang:1.15-alpine AS final
LABEL author="Iaroslav Akimov"

# line 1: making glibc's binaries link with libc.musl
# line 2: create config directory
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 \
	&& mkdir -p /etc/cacheman

COPY --from=builder /usr/src/cacheman/cacheman /usr/bin/cacheman
COPY --from=builder /usr/src/cacheman/config.json /etc/cacheman

# server http port
EXPOSE 8080

# server replication port
EXPOSE 8000

VOLUME /etc/cacheman

ENTRYPOINT ["/usr/bin/cacheman"]
