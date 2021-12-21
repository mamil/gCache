module gCache

go 1.14

replace gCache/gcache => ./gcache

require (
	gCache/gcache v0.0.0-00010101000000-000000000000
	github.com/imroc/req v0.3.2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/viper v1.10.1
)
