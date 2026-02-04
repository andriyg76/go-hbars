package runtime

import "sync"

// ReflectUsageLevel controls how reflection is used when type switches don't cover a value.
const (
	ReflectUse ReflectUsageLevel = iota
	ReflectWarn
	ReflectError
)

// ReflectUsageLevel is the level for reflect fallback branches.
type ReflectUsageLevel int

var (
	reflectLevel   ReflectUsageLevel = ReflectWarn
	reflectLevelMu sync.RWMutex
)

// SetReflectUsageLevel sets the reflect usage level (USE, WARN, ERROR). Default is WARN.
func SetReflectUsageLevel(level ReflectUsageLevel) {
	reflectLevelMu.Lock()
	defer reflectLevelMu.Unlock()
	reflectLevel = level
}

// GetReflectUsageLevel returns the current reflect usage level.
func GetReflectUsageLevel() ReflectUsageLevel {
	reflectLevelMu.RLock()
	defer reflectLevelMu.RUnlock()
	return reflectLevel
}
