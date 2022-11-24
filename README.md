# config-watcher

`config-watcher` is a simple project to watch external data stores and detect changes to keys and signal change in the
values.

## Supported Datastore

- Consul

## Usage

```go
package main

// Import relevant packages
import (
	"github.com/bubunyo/config-watcher/common"
	"github.com/bubunyo/config-watcher/consul"
)


// Setup Datastore configurations
config := &consul.Config{
    Address: "localhost:8500", // Consul address
    Config: common.Config{
        PollInterval: 1 * time.Second, // Polling Interval
        CloseTimeout: 5 * time.Second, // Close Timeout
    },
}

// Create consul watcher
watcher, err := consul.NewWatcher(config)
if err != nil ...

// Watch for changes by calling Wact(context, key) which returns a chanl of type []byte.
// Values are sent down the chanl only when changes are detected
// You can also watch for multiple keys on the same watcher
go func() {
    for byteSlice := range watcher.Watch(context.Background(), "a/b/c/d") {
        // use byteSlice	
    }
}()
go func() {
    for byteSlice := range watcher.Watch(context.Background(), "test_key") {
        // use byteSlice	
    }
}()

// or watch for the same key are multiple places, by calling watch with the same 
// key multiple times

// Finally Close the watcher when shutting down
err := watcher.Close()
if err != nil ...
```

## Testing Consul Integration

Docker is a prerequisite to test consul integrations

1. From the project folder, setup docker containers 
```
docker-compose -f ./docker/docker-compose.yml up --detach
```
2. Once your containers have started, you can import some sample values 
```
$ curl --request PUT --data-binary "@./docker/sample1.json" http://127.0.0.1:8500/v1/kv/backend
$ curl --request PUT --data-binary "@./docker/sample2.json" http://127.0.0.1:8500/v1/kv/a/b/c/d
```
3. You can verify if your keys have been imported into consul by visiting `http://localhost:8500/ui/dc1/kv`
4. Run the test `TestConsulWatcher`, without `t.Skip()` to verify

## Todo

- Add Tests
- Add Documentations
- Export Metrics