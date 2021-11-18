module github.com/goes-funky/httprouter/zapdriver

go 1.17

require (
	github.com/goes-funky/httprouter v0.0.0-20211118180036-82957f41fe1a
	github.com/goes-funky/zapdriver v0.0.0-20211118165755-ca718bf849a6
	go.uber.org/zap v1.19.1
)

require (
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
)

replace github.com/goes-funky/httprouter => ../
