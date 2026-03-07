package components

import "github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"

type LockableComponents struct {
	IsLocked bool
}

func (e *LockableComponents) Type() ecs.ComponentType {
	return ecs.ComponentTypeLockable
}

func NewLockableComponent(isLocked bool) *LockableComponents {
	return &LockableComponents{
		IsLocked: isLocked,
	}
}
