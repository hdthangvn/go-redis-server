// Package command holds the "CPU work" of the server: the in-memory store
// and command dispatch. It has no notion of sockets, goroutines, or event
// loops, so any I/O strategy (thread pool, single-threaded, io-multiplexed)
// can call Handle the same way.

// Contains a map[string]string (like a locker system) protected by a sync.Mutex (lock),
// with a Handle(args) function that knows how to execute PING, SET, and GET commands.
package command

import (
	"strings"
	"sync"

	"go-redis-server/lecture4_io_multiplexing/internal/protocol"
)

var store = struct {
	sync.Mutex
	data map[string]string
}{data: make(map[string]string)}

func Handle(args []string) []byte {
	if len(args) == 0 {
		return protocol.NilReply
	}
	switch strings.ToUpper(args[0]) {
	case "PING":
		return protocol.EncodeSimpleString("PONG")
	case "SET":
		if len(args) < 3 {
			return protocol.EncodeError("ERR wrong number of arguments for 'SET'")
		}
		store.Lock()
		store.data[args[1]] = args[2]
		store.Unlock()
		return protocol.EncodeSimpleString("OK")
	case "GET":
		if len(args) < 2 {
			return protocol.EncodeError("ERR wrong number of arguments for 'GET'")
		}
		store.Lock()
		v, ok := store.data[args[1]]
		store.Unlock()
		if !ok {
			return protocol.NilReply
		}
		return protocol.EncodeBulkString(v)
	default:
		return protocol.EncodeError("ERR unknown command")
	}
}
