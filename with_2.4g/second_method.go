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
			if t0 > Ts {
				t0 = Ts
			}
			if SumOfTinState(states)+t0 >= Ts {
				remaining := Ts - SumOfTinState(states)
				if remaining > 0 {
					states = append(states, State{"Disconnect", remaining})
				}
				break
			}
			states = append(states, State{"Disconnect", t0})
		} else {
			t1 := InverseCDFExponential(rand.Float64(), expectedValueT1)
			if t1 > Ts {
				t1 = Ts
			}
			if SumOfTinState(states)+t1 >= Ts {
				remaining := Ts - SumOfTinState(states)
				if remaining > 0 {
					states = append(states, State{"Connect", remaining})
				}
				break
			}
			states = append(states, State{"Connect", t1})
		}
		currentState = NextState(currentState)
	}
	return states
}

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

func Pareto(alpha, xm float64) float64 {
	u := rand.Float64()
	return xm / math.Pow(u, 1.0/alpha)
}

// === New Function ===
// MeasureBandwidth checks if total bandwidth in an iteration exceeds 500
// Returns 1 if yes, 0 if no
func MeasureBandwidth(states []State, wifiSpeed, mobileSpeed float64) int {
	totalb := 0.0
	for _, s := range states {
		if s.state == "Connect" {
			totalb += (wifiSpeed + mobileSpeed)
		} else if s.state == "Disconnect" {
			totalb += mobileSpeed
		}
	}
	avgBandwidth := totalb / float64(len(states))
	if avgBandwidth > 40 {
		return 1
	}
	return 0
}

func main() {
	rand.Seed(time.Now().UnixNano())

	iterations := 50000
	expectedValueSession := 50.0
	expectedValueT0 := 500.0
	expectedValueT1 := 50.0
	wifiSpeed := 60.0   // Mbps
	mobileSpeed := 40.0 // Mbps
	alpha := 25.0 / 3.0 // shape
	xm := 220.0         // minimum file size
	AvgRemainFailure := 0.0
	deadline := 0

	results := []int{} // to store bandwidth flags for each iteration

	for i := 1; i <= iterations; i++ {
		Ts := expectedValueSession
		fmt.Printf("\n================== Iteration %d ==================\n", i)
		fmt.Printf("\nSession Time: %.2f seconds\n", Ts)

		fileSizeMB := Pareto(alpha, xm)
		totalFileSizeMb := fileSizeMB * 8
		remainingSize = totalFileSizeMb
		fmt.Printf("File Size: %.2f MB (%.2f Mb)\n", fileSizeMB, totalFileSizeMb)
		fmt.Println("=======================================================================")

		// Generate the connect/disconnect sequence
		initialState := InitState(expectedValueT0, expectedValueT1)
		states := GenerateBand(initialState, expectedValueT0, expectedValueT1, Ts)

		// === Bandwidth check ===
		flag := MeasureBandwidth(states, wifiSpeed, mobileSpeed)
		results = append(results, flag)
		fmt.Printf("Iteration %d Bandwidth flag: %d\n", i, flag)

		// Run simulation
		SimulateDownload(states, wifiSpeed, mobileSpeed)

		// Final status
		fmt.Println("=======================================================================")
		if remainingSize > 0 {
			fmt.Println("File Download Incomplete")
			fmt.Printf("Total Downloaded: %.2f Mb (%.2f MB)\n", totalFileSizeMb-remainingSize, (totalFileSizeMb-remainingSize)/8)
			fmt.Printf("Remaining: %.2f Mb (%.2f MB)\n", remainingSize, remainingSize/8)
			AvgRemainFailure = AvgRemainFailure + remainingSize
			deadline = deadline + 1
		} else {
			fmt.Println("File Download Complete")
		}
	}

	if AvgRemainFailure > 0 {
		fmt.Println("=======================================================================")
		AvgRemainFailure = AvgRemainFailure / float64(iterations)
		deadlineRatio := float64(deadline) / float64(iterations)
		fmt.Printf("Average Remaining Size after %d iterations: %.2f Mb (%.2f MB)\n", iterations, AvgRemainFailure, AvgRemainFailure/8)
		fmt.Printf("Deadline miss(%d) ratio after %d iterations: %.5f\n", deadline, iterations, deadlineRatio)
	}
	// Calculate ratio of 1s in results

	sum := 0
	for _, v := range results {
		sum += v
	}
	ratio := float64(sum) / float64(len(results))

	// Print the flags array
	fmt.Println("=======================================================================")
	// fmt.Println("Bandwidth Flags per Iteration:", results)
	fmt.Printf("Ratio of iterations with bandwidth > 40: %.2f\n", ratio)

}
