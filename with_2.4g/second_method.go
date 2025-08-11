package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// State represents a network state and its duration (T)
type State struct {
	state string
	T     float64
}

var remainingSize float64

// DownloadFile simulates downloading during "Connect" periods
func DownloadFile(value float64, speed float64) float64 {
	fmt.Printf("Downloading for %.2f seconds at %.2f Mbps\n", value, speed)
	remainingSize -= (speed * value)
	return remainingSize
}

// SumOfTinState sums the T values in a slice of State structs
func SumOfTinState(states []State) float64 {
	sum := 0.0
	for _, s := range states {
		sum += s.T
	}
	return sum
}

// GenerateBand generates alternating "Connect" and "Disconnect" periods until total session time is reached
func GenerateBand(initialState string, expectedValueT0, expectedValueT1, Ts float64) []State {
	states := []State{}
	currentState := initialState

	for {
		if currentState == "disconnect" {
			t0 := InverseCDFExponential(rand.Float64(), expectedValueT0)
			states = append(states, State{"Disconnect", t0})
			if SumOfTinState(states) >= Ts {
				break
			}
		} else {
			t1 := InverseCDFExponential(rand.Float64(), expectedValueT1)
			if SumOfTinState(states)+t1 >= Ts {
				downloadable := Ts - SumOfTinState(states)
				states = append(states, State{"Connect", downloadable})
				break
			}
			states = append(states, State{"Connect", t1})
		}
		currentState = NextState(currentState)
	}
	return states
}

// SimulateDownload runs through each state and downloads if connected
// SimulateDownload runs through each state and downloads depending on connection type
func SimulateDownload(states []State, wifiSpeed, mobileSpeed float64) {
	for i, s := range states {
		fmt.Printf("State %d: %s, T: %.2f\n", i+1, s.state, s.T)
		if s.state == "Connect" {
			// WiFi + 4G/5G
			totalSpeed := wifiSpeed + mobileSpeed
			DownloadFile(s.T, totalSpeed)
		} else {
			// Only 4G/5G
			DownloadFile(s.T, mobileSpeed)
		}
		if remainingSize <= 0 {
			return
		}
	}
}

// Helper functions
func InverseCDFExponential(u, val float64) float64 {
	return (-val) * math.Log(1-u)
}

func NextState(state string) string {
	if state == "disconnect" {
		return "connect"
	}
	return "disconnect"
}

func InitState(expectedValueT0, expectedValueT1 float64) string {
	p0 := expectedValueT0 / (expectedValueT1 + expectedValueT0)
	u := rand.Float64()
	if u <= p0 {
		return "disconnect"
	}
	return "connect"
}

func main() {
	rand.Seed(time.Now().UnixNano())

	iterations := 3
	expectedValueSession := 150.0
	expectedValueT0 := 60.0
	expectedValueT1 := 40.0
	wifiSpeed := 200.0   // Mbps
	mobileSpeed := 50.0  // Mbps
	fileSizeMB := 1000.0 // MB

	for i := 1; i <= iterations; i++ {

		Ts := rand.Float64() * expectedValueSession
		fmt.Printf("\n================== Iteration %d ==================\n", i)
		fmt.Printf("\nSession Time: %.2f seconds\n", Ts)

		totalFileSizeMb := fileSizeMB * 8
		remainingSize = totalFileSizeMb
		fmt.Printf("File Size: %.2f MB (%.2f Mb)\n", fileSizeMB, totalFileSizeMb)
		fmt.Println("=======================================================================")

		// Speed of our single network

		// Generate the connect/disconnect sequence
		initialState := InitState(expectedValueT0, expectedValueT1)
		states := GenerateBand(initialState, expectedValueT0, expectedValueT1, Ts)

		// Show generated states
		fmt.Println("Generated States:")
		for _, s := range states {
			fmt.Printf("(%s, %.2f)\n", s.state, s.T)
		}
		fmt.Println("=======================================================================")

		// Run simulation
		SimulateDownload(states, wifiSpeed, mobileSpeed)

		// Final status
		fmt.Println("=======================================================================")
		if remainingSize > 0 {
			fmt.Println("File Download Incomplete")
			fmt.Printf("Total Downloaded: %.2f Mb (%.2f MB)\n", totalFileSizeMb-remainingSize, (totalFileSizeMb-remainingSize)/8)
			fmt.Printf("Remaining: %.2f Mb (%.2f MB)\n", remainingSize, remainingSize/8)
		} else {
			fmt.Println("\nFile Download Complete")

		}
	}
}
