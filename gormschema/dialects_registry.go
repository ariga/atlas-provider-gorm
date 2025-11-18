package gormschema

import (
	"fmt"
	"slices"
	"sync"

	"gorm.io/gorm"
)

type DialectFactory func(*Loader) (gorm.Dialector, error)

var (
	dialectorRegistry   = map[string]DialectFactory{}
	dialectorRegistryMu sync.RWMutex
)

// registers a new dialector under the provided name.
func RegisterDialector(name string, factory DialectFactory) {
	if name == "" {
		panic("gormschema: register dialector: empty name")
	}
	if factory == nil {
		panic(fmt.Sprintf("gormschema: register dialector %q: nil factory", name))
	}
	dialectorRegistryMu.Lock()
	defer dialectorRegistryMu.Unlock()
	if _, ok := dialectorRegistry[name]; ok {
		panic(fmt.Sprintf("gormschema: register dialector %q: already registered", name))
	}
	dialectorRegistry[name] = factory
}

func RegisteredDialects() []string {
	dialectorRegistryMu.RLock()
	defer dialectorRegistryMu.RUnlock()
	names := make([]string, 0, len(dialectorRegistry))
	for name := range dialectorRegistry {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (l *Loader) resolveDialector() (gorm.Dialector, error) {
	dialectorRegistryMu.RLock()
	factory := dialectorRegistry[l.dialect]
	dialectorRegistryMu.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("unsupported dialect %q (import the relevant gormschema/dialect/<name> package to enable it)", l.dialect)
	}
	return factory(l)
}
