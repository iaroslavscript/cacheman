module github.com/iaroslavscript/cacheman/lib/simplereplication

go 1.15

replace github.com/iaroslavscript/cacheman/lib/config => ../config

replace github.com/iaroslavscript/cacheman/lib/sdk => ../sdk

require (
	github.com/iaroslavscript/cacheman/lib/config v0.0.0-00010101000000-000000000000
	github.com/iaroslavscript/cacheman/lib/sdk v0.1.0
)
