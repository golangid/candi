# Example

## Create delivery handler

```go
package workerhandler

import (
	"context"
	"time"

	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
)

// RedisHandler struct
type RedisHandler struct {
}

// NewRedisHandler constructor
func NewRedisHandler() *RedisHandler {
	return &RedisHandler{}
}

// MountHandlers return group map topic key to handler func
func (h *RedisHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	
	group.Add("scheduled-push-notif", h.handleScheduledPushNotif)
	group.Add("heavy-push", h.handleHeavyPush)
}

func (h *RedisHandler) handleScheduledPushNotif(ctx context.Context, message []byte) error {
	// process usecase
	logger.LogIf("success handling message: %s", string(message))
	return err
}

func (h *RedisHandler) handleHeavyPush(ctx context.Context, message []byte) error {
	logger.LogI("start heavy push")
	time.Sleep(30 * time.Second)
	logger.LogI("heavy push done")
	return nil
}
```

## Register in module

```go
package examplemodule

import (

	"example.service/internal/modules/examplemodule/delivery/workerhandler"

	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

type Module struct {
	// ...another delivery handler
	workerHandlers map[types.Worker]interfaces.WorkerHandler
}

func NewModules(deps dependency.Dependency) *Module {
	return &Module{
		workerHandlers: map[types.Worker]interfaces.WorkerHandler{
			// ...another worker handler
			// ...
			types.RedisSubscriber: workerhandler.NewRedisHandler(),
		},
	}
}

// ...another method
```

## Add job in each usecase module

```go
package usecase

import (
	"context"
	"time"

	"github.com/golangid/candi/candihelper"
)

func (uc *usecaseImpl) someUsecase(ctx context.Context) {
	// scheduled exec to "scheduled-push-notif" handler after 5 minutes from now
	key := candihelper.BuildRedisPubSubKeyTopic("scheduled-push-notif", map[string]string{"message": "hello"})
	uc.cache.Set(ctx, key, "ok", 5*time.Minute)
}
```