module github.com/iaroslavscript/cacheman

go 1.15

replace github.com/iaroslavscript/cacheman/lib/config => ./lib/config

replace github.com/iaroslavscript/cacheman/lib/sdk => ./lib/sdk // indirect

replace github.com/iaroslavscript/cacheman/lib/server => ./lib/server

replace github.com/iaroslavscript/cacheman/lib/simplecache => ./lib/simplecache

replace github.com/iaroslavscript/cacheman/lib/simplereplication => ./lib/simplereplication

replace github.com/iaroslavscript/cacheman/lib/simplescheduler => ./lib/simplescheduler

require (
	github.com/iaroslavscript/cacheman/lib/config v0.0.0-00010101000000-000000000000
	github.com/iaroslavscript/cacheman/lib/server v0.0.0-00010101000000-000000000000
	github.com/iaroslavscript/cacheman/lib/simplecache v0.0.0-00010101000000-000000000000
	github.com/iaroslavscript/cacheman/lib/simplereplication v0.0.0-00010101000000-000000000000
	github.com/iaroslavscript/cacheman/lib/simplescheduler v0.0.0-00010101000000-000000000000
)
