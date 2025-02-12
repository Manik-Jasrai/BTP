package main

import (
    "fmt"
    "math"
    "math/rand"
    "time"
)

// AlgorithmType defines which algorithm to use
type AlgorithmType int

const (
    PartialAllocation AlgorithmType = iota // Original algorithm with partial allocations
    GPG                                    // Generalized Perturbed-Greedy
)

type Advertiser struct {
    ID            int
    InitialBudget float64
    Budget        float64
    Y             float64 // Random U[0,1] sample
}

type AdSystem struct {
    advertisers    map[int]*Advertiser
    beta          float64
    time          int
    algorithmType AlgorithmType
    availableI    map[int]bool // Used for GPG algorithm
}

// NewAdSystem creates a new ad system with specified algorithm type
func NewAdSystem(advertisers []*Advertiser, beta float64, algType AlgorithmType) *AdSystem {
    advertiserMap := make(map[int]*Advertiser)
    availableI := make(map[int]bool)
    
    for _, adv := range advertisers {
        adv.Y = rand.Float64()
        advertiserMap[adv.ID] = adv
        availableI[adv.ID] = true
    }
    
    return &AdSystem{
        advertisers:    advertiserMap,
        beta:          beta,
        time:          0,
        algorithmType: algType,
        availableI:    availableI,
    }
}

func (a *AdSystem) calculateG(t int) float64 {
    return math.Exp(a.beta * float64(t-1))
}

func (a *AdSystem) calculateGForY(y float64) float64 {
    return math.Exp(a.beta * (y - 1))
}

// ProcessNewArrival handles new bid arrivals using the selected algorithm
func (a *AdSystem) ProcessNewArrival(bids map[int]float64) interface{} {
    a.time++
    
    switch a.algorithmType {
    case PartialAllocation:
        return a.processPartialAllocation(bids)
    case GPG:
        return a.processGPG(bids)
    default:
        return nil
    }
}

// processPartialAllocation implements the original algorithm with partial allocations
func (a *AdSystem) processPartialAllocation(bids map[int]float64) map[int]float64 {
    allocations := make(map[int]float64)
    delta := 0.0
    
    // Get available advertisers
    available := make([]int, 0)
    for id, adv := range a.advertisers {
        if adv.Budget > 0 {
            available = append(available, id)
        }
    }
    
    gt := a.calculateG(a.time)
    
    // Main allocation loop
    for delta < 1 && len(available) > 0 {
        // Find best advertiser
        var bestAdv *Advertiser
        maxValue := -1.0
        
        for _, id := range available {
            if bids[id] <= 0 {
                continue
            }
            adv := a.advertisers[id]
            value := bids[id] * (1 - gt*adv.Y)
            if value > maxValue {
                maxValue = value
                bestAdv = adv
            }
        }
        
        if bestAdv == nil {
            break
        }
        
        // Calculate allocation
        deltaI := math.Min(1-delta, bestAdv.Budget/bids[bestAdv.ID])
        if deltaI > 0 {
            allocations[bestAdv.ID] = deltaI
            bestAdv.Budget -= bids[bestAdv.ID] * deltaI
            delta += deltaI
        }
        
        // Update available advertisers
        newAvailable := make([]int, 0)
        for _, id := range available {
            if a.advertisers[id].Budget > 0 {
                newAvailable = append(newAvailable, id)
            }
        }
        available = newAvailable
    }
    
    return allocations
}

// processGPG implements the Generalized Perturbed-Greedy algorithm
func (a *AdSystem) processGPG(bids map[int]float64) int {
    maxValue := -1.0
    bestAdvertiserID := -1
    
    // Find i* = argmax(bi,t(1 - g(yi)))
    for id := range a.availableI {
        if !a.availableI[id] || bids[id] <= 0 {
            continue
        }
        
        adv := a.advertisers[id]
        if adv.Budget < bids[id] {
            a.availableI[id] = false
            continue
        }
        
        value := bids[id] * (1 - a.calculateGForY(adv.Y))
        if value > maxValue {
            maxValue = value
            bestAdvertiserID = id
        }
    }
    
    // Update budget if match found
    if bestAdvertiserID != -1 {
        adv := a.advertisers[bestAdvertiserID]
        adv.Budget -= bids[bestAdvertiserID]
        
        if adv.Budget < 0.01 {
            a.availableI[bestAdvertiserID] = false
        }
    }
    
    return bestAdvertiserID
}

func generateTestData(numAdvertisers int, numArrivals int) ([]*Advertiser, []map[int]float64) {
    // Create advertisers with different budgets
    advertisers := make([]*Advertiser, numAdvertisers)
    for i := 0; i < numAdvertisers; i++ {
        budget := 100.0 + rand.Float64()*900.0 // Random budgets between 100-1000
        advertisers[i] = &Advertiser{
            ID:            i + 1,
            InitialBudget: budget,
            Budget:        budget,
        }
    }
    
    // Generate sequence of bids
    arrivals := make([]map[int]float64, numArrivals)
    for i := 0; i < numArrivals; i++ {
        bids := make(map[int]float64)
        for j := 1; j <= numAdvertisers; j++ {
            if rand.Float64() < 0.8 { // 80% chance of bidding
                bids[j] = 10.0 + rand.Float64()*40.0 // Random bids between 10-50
            }
        }
        arrivals[i] = bids
    }
    
    return advertisers, arrivals
}

func main() {
    rand.Seed(time.Now().UnixNano())
    
    // Test both algorithms
    algorithms := []struct {
        name AlgorithmType
        desc string
    }{
        {PartialAllocation, "Partial Allocation Algorithm"},
        {GPG, "Generalized Perturbed-Greedy Algorithm"},
    }
    
    for _, alg := range algorithms {
        fmt.Printf("\n=== Testing %s ===\n", alg.desc)
        
        // Generate test data
        advertisers, arrivals := generateTestData(5, 10)
        
        // Create allocation system with specified algorithm
        system := NewAdSystem(advertisers, 0.5, alg.name)
        
        // Process each arrival
        for i, bids := range arrivals {
            fmt.Printf("\nArrival %d:\n", i+1)
            fmt.Printf("Bids: %v\n", bids)
            
            result := system.ProcessNewArrival(bids)
            
            // Print results based on algorithm type
            switch alg.name {
            case PartialAllocation:
                fmt.Printf("Allocations: %v\n", result.(map[int]float64))
            case GPG:
                fmt.Printf("Matched to Advertiser: %d\n", result.(int))
            }
            
            // Print remaining budgets
            fmt.Println("Remaining budgets:")
            for _, adv := range advertisers {
                fmt.Printf("Advertiser %d: %.2f (y=%.3f)\n", 
                    adv.ID, adv.Budget, adv.Y)
            }
        }
    }
}
