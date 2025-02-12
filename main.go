package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Advertiser struct {
	ID            int
	InitialBudget float64
	Budget        float64
	Y             float64 // Random U[0,1] sample
}

type AdAllocation struct {
	advertisers map[int]*Advertiser
	beta        float64
	time        int
}

func NewAdAllocation(advertisers []*Advertiser, beta float64) *AdAllocation {
	advertiserMap := make(map[int]*Advertiser)
	for _, adv := range advertisers {
		// Generate random y value for each advertiser
		adv.Y = rand.Float64()
		advertiserMap[adv.ID] = adv
	}

	return &AdAllocation{
		advertisers: advertiserMap,
		beta:        beta,
		time:        0,
	}
}

func (a *AdAllocation) calculateG(t int) float64 {
	return math.Exp(a.beta * float64(t-1))
}

func (a *AdAllocation) ProcessNewArrival(bids map[int]float64) map[int]float64 {
	a.time++
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

func generateTestData(numAdvertisers int, numArrivals int) (
	[]*Advertiser,
	[]map[int]float64,
) {
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
			// Random bids between 10-50
			if rand.Float64() < 0.8 { // 80% chance of bidding
				bids[j] = 10.0 + rand.Float64()*40.0
			}
		}
		arrivals[i] = bids
	}

	return advertisers, arrivals
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Generate test data
	advertisers, arrivals := generateTestData(5, 10)

	// Create allocation system
	allocation := NewAdAllocation(advertisers, 0.5)

	// Process each arrival
	for i, bids := range arrivals {
		fmt.Printf("\nArrival %d:\n", i+1)
		fmt.Printf("Bids: %v\n", bids)
		result := allocation.ProcessNewArrival(bids)
		fmt.Printf("Allocations: %v\n", result)

		// Print remaining budgets
		fmt.Println("Remaining budgets:")
		for _, adv := range advertisers {
			fmt.Printf("Advertiser %d: %.2f\n", adv.ID, adv.Budget)
		}
	}
}
