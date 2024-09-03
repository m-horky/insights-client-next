package collectors

import (
	"github.com/m-horky/insights-client-next/app"
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
func GetCollector(name string) (*Collector, app.HumanError) {
	for _, collector := range GetCollectors() {
		if collector.Name == name {
			return collector, nil
		}
	}
	return nil, app.NewError(ErrNoCollector, nil, "Collector not found.")
}

func (c *Collector) Run() app.HumanError {
	return app.NewError(nil, nil, "Not implemented")
}
