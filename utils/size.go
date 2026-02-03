package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

type SimulatorInfo struct {
	Name       string
	Size       int64
	UDID       string
	IsOrphaned bool
}

func CalculateDirSize(root string) int64 {
	var totalSize int64
	semaphore := make(chan struct{}, 100) // Boost concurrency

	var walk func(string)
	walk = func(path string) {
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		info, err := os.Lstat(path)
		if err != nil {
			return
		}
		if !info.IsDir() {
			atomic.AddInt64(&totalSize, info.Size())
			return
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return
		}
		for _, entry := range entries {
			walk(filepath.Join(path, entry.Name()))
		}
	}

	walk(root)

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	return totalSize
}

type Device struct {
	Name        string `json:"name"`
	UDID        string `json:"udid"`
	Runtime     string `json:"runtime"`
	State       string `json:"state"`
	IsAvailable bool   `json:"isAvailable"`
}

type Devices struct {
	Devices map[string][]Device `json:"devices"`
}

type SimNameInfo struct {
	Name       string
	IsOrphaned bool
}

func GetSimFriendlyNames() map[string]SimNameInfo {
	out, err := exec.Command("xcrun", "simctl", "list", "--json", "devices").Output()
	if err != nil {
		fmt.Println("Failed to run simctl:", err)
		return nil
	}

	var sims Devices
	json.Unmarshal(out, &sims)

	nameMap := make(map[string]SimNameInfo)
	for runtime, devices := range sims.Devices {
		simplifiedRuntime := strings.TrimPrefix(runtime, "com.apple.CoreSimulator.SimRuntime.")
		parts := strings.SplitN(simplifiedRuntime, "-", 2)
		if len(parts) == 2 {
			platform := parts[0]
			version := strings.ReplaceAll(parts[1], "-", ".")
			simplifiedRuntime = fmt.Sprintf("%s %s", platform, version)
		}
		for _, dev := range devices {
			displayName := fmt.Sprintf("%s (%s)", dev.Name, simplifiedRuntime)
			if !dev.IsAvailable {
				displayName = fmt.Sprintf("Orphaned %s (%s)", dev.Name, simplifiedRuntime)
			}
			nameMap[dev.UDID] = SimNameInfo{
				Name:       displayName,
				IsOrphaned: !dev.IsAvailable,
			}
		}
	}
	return nameMap
}

func ListSimulators(showAll, showCriticalOnly bool, thresholdGB int, outputFile string, cleanInteractive, dryRun, forceClean, summaryOnly bool) {
	sims, totalSize, biggestSim := gatherSimulators(showAll, showCriticalOnly)

	criticalSims, normalSims := splitSimulators(sims, thresholdGB)

	// Count orphaned simulators
	orphanedCount := 0
	for _, sim := range sims {
		if sim.IsOrphaned {
			orphanedCount++
		}
	}

	var out strings.Builder

	printSimulators(criticalSims, normalSims, &out)
	printSummary(totalSize, biggestSim, len(criticalSims), len(normalSims), orphanedCount, &out)

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(out.String()), 0644)
		if err != nil {
			color.Red("❌ Failed to save report: %v", err)
		} else {
			color.Green("✅ Report saved to %s", outputFile)
		}
	}

	if cleanInteractive {
		interactiveClean(criticalSims)
	}
}

func gatherSimulators(showAll, showCriticalOnly bool) ([]SimulatorInfo, int64, SimulatorInfo) {
	base := ExpandPath("~/Library/Developer/CoreSimulator/Devices")
	entries, err := os.ReadDir(base)
	if err != nil {
		color.Red("Failed to read simulators directory")
		return nil, 0, SimulatorInfo{}
	}

	nameMap := GetSimFriendlyNames()
	var sims []SimulatorInfo
	var totalSize int64
	var biggestSim SimulatorInfo

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = "  Calculating Simulator Sizes…"
	s.Start()

	for _, e := range entries {
		uuid := e.Name()
		dataPath := filepath.Join(base, uuid, "data")
		size := CalculateDirSize(dataPath)

		if !showAll && size < (10<<20) {
			continue
		}
		if showCriticalOnly && size < (3<<30) {
			continue
		}

		var friendlyName string
		var isOrphaned bool
		if info, found := nameMap[uuid]; found {
			friendlyName = info.Name
			isOrphaned = info.IsOrphaned
		} else {
			friendlyName = fmt.Sprintf("Orphaned (%s)", uuid)
			isOrphaned = true
		}

		sim := SimulatorInfo{
			Name:       friendlyName,
			Size:       size,
			UDID:       uuid,
			IsOrphaned: isOrphaned,
		}

		sims = append(sims, sim)
		totalSize += size

		if sim.Size > biggestSim.Size {
			biggestSim = sim
		}
	}

	s.Stop()
	return sims, totalSize, biggestSim
}

func splitSimulators(sims []SimulatorInfo, thresholdGB int) ([]SimulatorInfo, []SimulatorInfo) {
	var critical []SimulatorInfo
	var normal []SimulatorInfo

	if thresholdGB > 0 {
		for _, sim := range sims {
			if sim.Size >= int64(thresholdGB)<<30 {
				critical = append(critical, sim)
			}
		}
	} else {
		for _, sim := range sims {
			if sim.Size >= 3<<30 {
				critical = append(critical, sim)
			} else {
				normal = append(normal, sim)
			}
		}
	}

	sort.Slice(critical, func(i, j int) bool {
		return critical[i].Size > critical[j].Size
	})
	sort.Slice(normal, func(i, j int) bool {
		return normal[i].Size > normal[j].Size
	})

	return critical, normal
}

func printSimulators(criticalSims, normalSims []SimulatorInfo, out *strings.Builder) {
	fmt.Println()
	color.Cyan("📱 Xcode Simulators")
	out.WriteString("\n📱 Xcode Simulators\n\n")

	if len(criticalSims) > 0 {
		color.Red("CRITICAL Simulators:")
		out.WriteString("CRITICAL Simulators:\n")
		for _, sim := range criticalSims {
			line := fmt.Sprintf("CRITICAL: %s — %.2f GB\n", sim.Name, float64(sim.Size)/(1<<30))
			if sim.IsOrphaned {
				color.New(color.FgHiYellow, color.Bold).Print("⚠️  " + line)
				out.WriteString("⚠️  " + line)
			} else if sim.Size > 10<<30 {
				color.New(color.FgHiRed, color.Bold).Print("🔥 " + line)
				out.WriteString("🔥 " + line)
			} else {
				color.New(color.FgRed, color.Bold).Print(line)
				out.WriteString(line)
			}
		}
		fmt.Println()
		out.WriteString("\n")
	}

	if len(normalSims) > 0 {
		color.Cyan("Normal Simulators:")
		out.WriteString("Normal Simulators:\n")
		for _, sim := range normalSims {
			line := fmt.Sprintf("%s — %.2f GB\n", sim.Name, float64(sim.Size)/(1<<30))
			if sim.IsOrphaned {
				color.Yellow("⚠️  " + line)
				out.WriteString("⚠️  " + line)
			} else {
				color.Green(line)
				out.WriteString(line)
			}
		}
		fmt.Println()
		out.WriteString("\n")
	}
}

func printSummary(totalSize int64, biggestSim SimulatorInfo, criticalCount, normalCount, orphanedCount int, out *strings.Builder) {
	fmt.Println()

	if biggestSim.Size > 10<<30 {
		color.Yellow("📈 Biggest Simulator: 🔥 %s — %.2f GB", biggestSim.Name, float64(biggestSim.Size)/(1<<30))
		out.WriteString(fmt.Sprintf("📈 Biggest Simulator: 🔥 %s — %.2f GB\n", biggestSim.Name, float64(biggestSim.Size)/(1<<30)))
	} else {
		color.Yellow("📈 Biggest Simulator: %s — %.2f GB", biggestSim.Name, float64(biggestSim.Size)/(1<<30))
		out.WriteString(fmt.Sprintf("📈 Biggest Simulator: %s — %.2f GB\n", biggestSim.Name, float64(biggestSim.Size)/(1<<30)))
	}

	color.Green("\n📱 Total Simulator Space Used: %.2f GB", float64(totalSize)/(1<<30))
	out.WriteString(fmt.Sprintf("\n📱 Total Simulator Space Used: %.2f GB\n", float64(totalSize)/(1<<30)))

	fmt.Println()
	if orphanedCount > 0 {
		color.Cyan("Summary: %d Critical Sims, %d Normal Sims, %d Orphaned (runtime unavailable)", criticalCount, normalCount, orphanedCount)
		out.WriteString(fmt.Sprintf("Summary: %d Critical Sims, %d Normal Sims, %d Orphaned (runtime unavailable)\n", criticalCount, normalCount, orphanedCount))
	} else {
		color.Cyan("Summary: %d Critical Sims, %d Normal Sims", criticalCount, normalCount)
		out.WriteString(fmt.Sprintf("Summary: %d Critical Sims, %d Normal Sims\n", criticalCount, normalCount))
	}
}

func interactiveClean(sims []SimulatorInfo) {
	if len(sims) == 0 {
		color.Yellow("No simulators to clean.")
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for _, sim := range sims {
		fmt.Printf("🧹 Delete %s (%.2f GB)? [y/N]: ", sim.Name, float64(sim.Size)/(1<<30))
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))

		if resp == "y" {
			color.Red("⚡ Deleting %s...", sim.Name)
			exec.Command("xcrun", "simctl", "delete", sim.UDID).Run()
		}
	}
}

func ListRuntimes() {
	out, err := exec.Command("xcrun", "simctl", "list", "--json", "runtimes").Output()
	if err != nil {
		color.Red("Failed to run simctl list runtimes: %v", err)
		return
	}

	var result struct {
		Runtimes []struct {
			Name         string `json:"name"`
			Identifier   string `json:"identifier"`
			Version      string `json:"version"`
			BuildVersion string `json:"buildversion"`
			IsAvailable  bool   `json:"isAvailable"`
		} `json:"runtimes"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		color.Red("Failed to parse runtimes JSON: %v", err)
		return
	}

	fmt.Println()
	color.Cyan("🧩 Installed Xcode Runtimes")
	fmt.Println()

	for _, runtime := range result.Runtimes {
		if runtime.IsAvailable {
			fmt.Printf("🛠️  %s (%s)\n", runtime.Name, runtime.Version)
		}
	}
}

func isRoot() bool {
	return os.Geteuid() == 0
}

func RemoveRuntime(targetName string, force bool) {
	out, err := exec.Command("xcrun", "simctl", "list", "--json", "runtimes").Output()
	if err != nil {
		color.Red("Failed to run simctl list runtimes: %v", err)
		return
	}

	var result struct {
		Runtimes []struct {
			Name        string `json:"name"`
			Identifier  string `json:"identifier"`
			IsAvailable bool   `json:"isAvailable"`
		} `json:"runtimes"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		color.Red("Failed to parse runtimes JSON: %v", err)
		return
	}

	var found string
	for _, runtime := range result.Runtimes {
		if strings.Contains(runtime.Name, targetName) && runtime.IsAvailable {
			found = runtime.Identifier
			break
		}
	}

	if found == "" {
		color.Red("❌ Could not find runtime matching: %s", targetName)
		return
	}

	color.Yellow("⚡ Deleting runtime: %s (%s)", targetName, found)

	// Check for force or root
	if !force && !isRoot() {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("This action requires sudo permission. Continue? [y/N]: ")
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))

		if resp != "y" {
			color.Red("❌ Cancelled runtime deletion.")
			return
		}
	}

	runtimeFolder := runtimeFolderFromIdentifier(found)
	cmd := exec.Command("sudo", "rm", "-rf", "/Library/Developer/CoreSimulator/Profiles/Runtimes/"+runtimeFolder)
	err = cmd.Run()
	if err != nil {
		color.Red("Failed to delete runtime: %v", err)
	} else {
		color.Green("✅ Successfully deleted runtime: %s", targetName)
		logDeletion(targetName)
	}
}

// Helper to get runtime folder name
func runtimeFolderFromIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	if len(parts) == 0 {
		return identifier
	}
	return parts[len(parts)-1] + ".simruntime"
}

func logDeletion(name string) {
	f, err := os.OpenFile(os.Getenv("HOME")+"/.xclean.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	logLine := fmt.Sprintf("%s Deleted runtime: %s\n", time.Now().Format(time.RFC3339), name)
	f.WriteString(logLine)
}
