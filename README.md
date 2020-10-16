# cacheman

## Intro

**CacheMan** is easy to use in-memory key-value storage server.
Server supports asynchronius replication. Current release works only in m*ster-mode,
sl*ve-mode is on the way.

## Usage

### Command line arguments

```                                                               
  -bind string
        http server bind address. (default "0.0.0.0:8080")
```

### Configuration file

Server **cacheman** uses configuration file `/etc/cacheman/config.json`
If file is absent the default settings are used.
And example of config.json together with default values is availible at 
[https://github.com/iaroslavscript/cacheman/blob/main/config.json](https://github.com/iaroslavscript/cacheman/blob/main/config.json)

* `bind_addr` string - http server bind address. (default *"0.0.0.0:8080"*)
* `expires_default_duration_sec` int - The default time for storing records in seconds (default *1800*)
* `replication_rotate_every_ms` int - The period of rotation replication log in milliseconds (default *1000*)
* `sheduler_del_expired_every_sec` int - The period of running deletion of expired records (default *60*)
* `sheduler_expired_queque_size` int - The maximum records for deleteion in queue (default *1000*)

### RestAPI

* `HEAD hostname:port/` - heath check-in. Responces with *200 OK* 
* `HEAD hostname:port/somekey` - Check key exists.
  * Responses with **200 OK** if key *somekey* exists.
  * Responses with **404 page not found** means key *somekey* is not present in storage or expired 
* `GET` - Lookup for the key.
  * Responses with **200 OK** (body contains value) if key *somekey* exists.
  * Responses with **404 page not found** means key *somekey* is not present in storage or expired
* `POST hostname:port/somekey` - Insert a new key or replace existed one. The value is taken from the body.
  * Recommended header `Content-Type` value is *text/plain; charset=utf-8*
  * Use header `X-Content-Expires-Sec` to set desired duration in seconds before key expires (default duration otherways)
  * Responses with **200 OK** if key-value was inserted
  * Responses with **400 Bad Request** in case of error
* `DELETE hostname:port/somekey` - Delete key from storage.
  * Responses with **200 OK** even if key *somekey* was not found

## Run in docker

### Building docker image
```
user@pc:~$ mkdir -p src/cacheman
user@pc:~$ cd src/cacheman
user@pc:~/src/cacheman$ git clone -q https://github.com/iaroslavscript/cacheman.git  
user@pc:~/src/cacheman$ sudo -s
root@pc:/home/user/src/cacheman# cd ~ 
root@kuhwpc:~# docker build --no-cache -t cacheman:0.1.1
```

Run docker container with exposed ports
```
root@kuhwpc:~# docker run -d -P --name=cacheman cacheman:0.1.1
root@kuhwpc:~# export SERVER_HTTP="$(docker container port cacheman 8080)"
```

To see container logs run
```
root@kuhwpc:~# docker logs -f cacheman
```

### Testing with curl
```
root@kuhwpc:~# curl -I $SERVER_HTTP
HTTP/1.1 200 OK
Date: Thu, 15 Oct 2020 15:37:10 GMT

root@kuhwpc:~# curl -i -d "{'x': 'y', 'z': 'q'}" -H X-Content-Expires-Sec:15 $SERVER_HTTP/keyA
HTTP/1.1 200 OK
Date: Thu, 15 Oct 2020 15:37:15 GMT
Content-Length: 0

root@kuhwpc:~# curl -i $SERVER_HTTP/keyA
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Thu, 15 Oct 2020 15:37:23 GMT
Content-Length: 22

{'x': 'y', 'z': 'q'}root@kuhwpc:~# 
root@kuhwpc:~# 
root@kuhwpc:~# sleep 30
root@kuhwpc:~# 
root@kuhwpc:~# curl -i $SERVER_HTTP/keyA
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 15 Oct 2020 15:38:03 GMT
Content-Length: 19

404 page not found
```

## Release notes

### v0.1.1
* bug fixes with rotating replication buckets
* more detailed log messages
* improved readme
