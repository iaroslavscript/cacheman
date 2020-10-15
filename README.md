# cacheman

## Intro

**CacheMan** is easy to use in-memory key-value storage.
The application by default is listening on `8080` for http traffic.

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

* `bind_addr` string - http server bind address. (default *"0.0.0.0:9080"*)
* `expires_default_duration_sec` int - The default time for storing records in seconds (default *3600*)
* `replication_rotate_every_ms` int - The period of rotation replication log in milliseconds (default *1000*)
* `sheduler_del_expired_every_sec` int - The period of running deletion of expired records (default *60*)
* `sheduler_expired_queque_size` int - The maximum records for deleteion in queue (default *1000*)

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

