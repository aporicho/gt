package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type pathStats struct {
	count    int
	lastUsed int64
}

func loadStats() map[string]pathStats {
	f, err := os.Open(statsFile)
	if err != nil {
		return make(map[string]pathStats)
	}
	defer f.Close()
	stats := make(map[string]pathStats)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		count, err1 := strconv.Atoi(parts[0])
		ts, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		stats[parts[2]] = pathStats{count: count, lastUsed: ts}
	}
	return stats
}

func saveStats(stats map[string]pathStats) {
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	f, err := os.Create(statsFile)
	if err != nil {
		return
	}
	defer f.Close()
	for _, k := range keys {
		s := stats[k]
		fmt.Fprintf(f, "%d\t%d\t%s\n", s.count, s.lastUsed, k)
	}
}

func frecencyWeight(lastUsed int64) float64 {
	hours := time.Since(time.Unix(lastUsed, 0)).Hours()
	switch {
	case hours < 1:
		return 4.0
	case hours < 24:
		return 2.0
	case hours < 24*7:
		return 1.0
	case hours < 24*30:
		return 0.5
	default:
		return 0.25
	}
}

func frecencyScore(s pathStats) float64 {
	return float64(s.count) * frecencyWeight(s.lastUsed)
}

func sortByFrecency(dirs []string) []string {
	stats := loadStats()
	sorted := make([]string, len(dirs))
	copy(sorted, dirs)
	sort.SliceStable(sorted, func(i, j int) bool {
		si := frecencyScore(stats[sorted[i]])
		sj := frecencyScore(stats[sorted[j]])
		if math.Abs(si-sj) > 1e-9 {
			return si > sj
		}
		return sorted[i] < sorted[j]
	})
	return sorted
}

func bumpStats(path string) {
	stats := loadStats()
	s := stats[path]
	s.count++
	s.lastUsed = time.Now().Unix()
	stats[path] = s
	saveStats(stats)
}
