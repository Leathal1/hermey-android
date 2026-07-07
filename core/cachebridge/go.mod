module github.com/Leathal1/hermey-android/core/cachebridge

go 1.23

require github.com/Leathal1/hermey-android/core/cache v0.0.0

require (
	go.etcd.io/bbolt v1.4.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace github.com/Leathal1/hermey-android/core/cache => ../cache
