module github.com/iaroslavscript/cacheman/lib/simplescheduler

replace github.com/iaroslavscript/cacheman/lib/sdk => ../sdk

replace github.com/iaroslavscript/cacheman/lib/config => ../config

go 1.15

require (
	github.com/iaroslavscript/cacheman/lib/config v0.0.0-00010101000000-000000000000
	github.com/iaroslavscript/cacheman/lib/sdk v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.8.0
)
