package tribft

import (
	"blockEmulator/params"
	"math/rand"
	"sync"
	"time"
)

// LayerType layer type
type LayerType int

const (
	LayerCity     LayerType = iota // City layer - transaction collection, local consensus
	LayerDistrict                  // District layer - cross-shard coordination, reputation aggregation
	LayerGlobal                    // Global layer - global state, reputation management
)

// LayerDelaySimulator layer delay simulator
// Simulates communication delays between different layers in TriBFT three-layer architecture
type LayerDelaySimulator struct {
	// Delay configuration (in milliseconds)
	cityToDistrictDelay   int64 // City -> District delay
	districtToGlobalDelay int64 // District -> Global delay
	interCityDelay        int64 // City-to-City delay (same shard)
	interDistrictDelay    int64 // District-to-District delay (cross-shard)

	// Delay variance (simulate network jitter)
	delayVariance float64

	// Random number generator
	rng  *rand.Rand
	lock sync.Mutex

	// Enable switch
	enabled bool
}

// NewLayerDelaySimulator creates a new layer delay simulator
func NewLayerDelaySimulator() *LayerDelaySimulator {
	lds := &LayerDelaySimulator{
		cityToDistrictDelay:   int64(params.ShardToCityBaseDelay),
		districtToGlobalDelay: int64(params.CityToGlobalBaseDelay),
		interCityDelay:        int64(params.IntraShardBaseDelay),
		interDistrictDelay:    int64(params.GlobalToShardBaseDelay),
		delayVariance:         0.3, // Default 30% variance
		rng:                   rand.New(rand.NewSource(time.Now().UnixNano())),
		enabled:               params.EnableLayerDelayLogging,
	}
	return lds
}

// SimulateCityToDistrict simulates City -> District layer communication delay
func (lds *LayerDelaySimulator) SimulateCityToDistrict() {
	if !lds.enabled {
		return
	}
	delay := lds.calculateDelay(lds.cityToDistrictDelay)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// SimulateDistrictToGlobal simulates District -> Global layer communication delay
func (lds *LayerDelaySimulator) SimulateDistrictToGlobal() {
	if !lds.enabled {
		return
	}
	delay := lds.calculateDelay(lds.districtToGlobalDelay)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// SimulateInterCity simulates City-to-City communication delay
func (lds *LayerDelaySimulator) SimulateInterCity() {
	if !lds.enabled {
		return
	}
	delay := lds.calculateDelay(lds.interCityDelay)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// SimulateInterDistrict simulates District-to-District communication delay (cross-shard)
func (lds *LayerDelaySimulator) SimulateInterDistrict() {
	if !lds.enabled {
		return
	}
	delay := lds.calculateDelay(lds.interDistrictDelay)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// calculateDelay calculates delay with random jitter
func (lds *LayerDelaySimulator) calculateDelay(baseDelay int64) int64 {
	lds.lock.Lock()
	defer lds.lock.Unlock()

	if lds.delayVariance == 0 {
		return baseDelay
	}

	// Add random variance
	variance := float64(baseDelay) * lds.delayVariance
	jitter := (lds.rng.Float64()*2 - 1) * variance
	result := float64(baseDelay) + jitter

	if result < 0 {
		return 0
	}
	return int64(result)
}

// SetEnabled enables or disables the simulator
func (lds *LayerDelaySimulator) SetEnabled(enabled bool) {
	lds.enabled = enabled
}

// IsEnabled returns whether the simulator is enabled
func (lds *LayerDelaySimulator) IsEnabled() bool {
	return lds.enabled
}

// String returns a readable representation of LayerType
func (lt LayerType) String() string {
	switch lt {
	case LayerCity:
		return "City"
	case LayerDistrict:
		return "District"
	case LayerGlobal:
		return "Global"
	default:
		return "Unknown"
	}
}
