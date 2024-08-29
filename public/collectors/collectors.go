package collectors

import (
	"errors"
	"fmt"
)

// Collector manages configuration required for data collection.
type Collector struct {
	Name        string
	Version     string
	Env         []string
	Exec        string
	ExecArgs    []string
	ContentType string
}

// GetCollectors gathers all available data collectors.
func GetCollectors() []*Collector {
	return []*Collector{
		GetAdvisorCollector(),
	}
}

// GetDefaultCollector returns the collector that should be run by default.
func GetDefaultCollector() *Collector {
	return GetAdvisorCollector()
}

// GetCollector filters available collectors by name.
func GetCollector(name string) (*Collector, error) {
	for _, collector := range GetCollectors() {
		if collector.Name == name {
			return collector, nil
		}
	}
	return nil, fmt.Errorf("collector '%s' not found", name)
}

func (c *Collector) Run() error {
	return errors.New("not implemented")
}
