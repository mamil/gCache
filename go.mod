module gcache

go 1.14

replace gcache/cache => ./cache

require (
	gcache/cache v0.0.0-00010101000000-000000000000
	github.com/alecthomas/assert v1.0.0
	github.com/imroc/req v0.3.2
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.10.1
)
