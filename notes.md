# Build Errors & How to fix them

## MacOS

### link: golang.org/x/net/internal/socket: invalid reference to syscall.recvmsg
Go Version: go version go1.27.2 darwin/arm64
Fix: Manually pull golang.org/x/net using `go get -u golang.org/x/net`
See issue https://github.com/google/broken-dns/issues/6

 
