package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/uwa/aisweep/internal/registry"
	"github.com/uwa/aisweep/internal/scanner"
	"github.com/uwa/aisweep/internal/server"
	"github.com/uwa/aisweep/internal/snapshot"
)

const usage = `aisweep - AI tool storage inventory

Usage:
  aisweep scan                  Run one scan, write snapshot
  aisweep serve [flags]         Start local dashboard
  aisweep path                  Print effective registry
  aisweep doctor                Check tool paths
  aisweep version
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(0)
	}
	switch os.Args[1] {
	case "scan":
		runScan(os.Args[2:])
	case "serve":
		runServe(os.Args[2:])
	case "path":
		runPath(os.Args[2:])
	case "doctor":
		runDoctor(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Println("aisweep dev")
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}

func runScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	quiet := fs.Bool("quiet", false, "suppress progress output")
	_ = fs.Parse(args)

	reg, err := registry.Load()
	if err != nil {
		die(err)
	}
	var out *os.File
	if !*quiet {
		out = os.Stdout
	}
	snap, err := scanner.Scan(reg, out)
	if err != nil {
		die(err)
	}
	if err := snapshot.Write(snap); err != nil {
		die(err)
	}
	fmt.Printf("snapshot saved: %s (%d entries, %s)\n",
		snap.ID, len(snap.Entries), snapshot.FormatBytes(totalSize(snap)))
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	host := fs.String("host", "127.0.0.1", "bind address")
	port := fs.Int("port", 7890, "listen port")
	scan := fs.Bool("scan", true, "run a scan on startup")
	interval := fs.Duration("interval", 0, "periodic scan interval (e.g. 1h, 15m); 0 disables")
	_ = fs.Parse(args)

	reg, err := registry.Load()
	if err != nil {
		die(err)
	}
	runScan := func() error {
		snap, err := scanner.Scan(reg, os.Stderr)
		if err != nil {
			return err
		}
		return snapshot.Write(snap)
	}
	if *scan {
		if err := runScan(); err != nil {
			fmt.Fprintln(os.Stderr, "scan error:", err)
		}
	}
	if *interval > 0 {
		go func() {
			t := time.NewTicker(*interval)
			defer t.Stop()
			for range t.C {
				if err := runScan(); err != nil {
					fmt.Fprintln(os.Stderr, "periodic scan error:", err)
				}
			}
		}()
		fmt.Fprintf(os.Stderr, "periodic scan every %s\n", *interval)
	}
	if err := server.Serve(*host, *port, runScan); err != nil {
		die(err)
	}
}

func runPath(args []string) {
	reg, err := registry.Load()
	if err != nil {
		die(err)
	}
	for _, t := range reg.Tools {
		for _, p := range t.AllPaths() {
			fmt.Printf("%-22s %-12s %-8s %s\n", t.ID, p.Category, p.Risk, p.Path)
		}
	}
}

func runDoctor(args []string) {
	reg, err := registry.Load()
	if err != nil {
		die(err)
	}
	for _, t := range reg.Tools {
		for _, p := range t.AllPaths() {
			if _, err := os.Stat(p.Path); err == nil {
				fmt.Printf("FOUND  %s\n", p.Path)
			} else {
				fmt.Printf("MISS   %s\n", p.Path)
			}
		}
	}
}

func totalSize(s *snapshot.Snapshot) int64 {
	var n int64
	for _, e := range s.Entries {
		n += e.SizeBytes
	}
	return n
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
