package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

var (
	opts = &runner.Options{}

	// perf profiling
	enableProfiling bool
	profileDir      string

	// hooks to run after command execution
	postRunHooks []func()
)

var rootCmd = &cobra.Command{
	Use:   "gowitness",
	Short: "A web screenshot and information gathering tool",
	Long:  ascii.Logo(),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if opts.Logging.Silence {
			log.EnableSilence()
		}

		if opts.Logging.Debug && !opts.Logging.Silence {
			log.EnableDebug()
			log.Debug("debug logging enabled")
		}

		if enableProfiling {
			ts := time.Now().Format("20060102-150405")
			profileDir = filepath.Join("profiles", ts)

			if err := os.MkdirAll(profileDir, 0o755); err != nil {
				return fmt.Errorf("could not create profile directory %q: %w", profileDir, err)
			}

			cpuPath := filepath.Join(profileDir, "cpu.pprof")
			memPath := filepath.Join(profileDir, "mem.pprof")
			tracePath := filepath.Join(profileDir, "trace.out")

			// cpu
			cpuFile, err := os.Create(cpuPath)
			if err != nil {
				return fmt.Errorf("could not create CPU profile file: %w", err)
			}
			if err := pprof.StartCPUProfile(cpuFile); err != nil {
				_ = cpuFile.Close()
				return fmt.Errorf("could not start CPU profile: %w", err)
			}
			postRunHooks = append(postRunHooks, func() {
				pprof.StopCPUProfile()
				_ = cpuFile.Close()
			})

			// memory
			postRunHooks = append(postRunHooks, func() {
				memFile, err := os.Create(memPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not create memory profile file: %v\n", err)
					return
				}
				defer memFile.Close()

				runtime.GC() // refresh heap statistics

				if err := pprof.WriteHeapProfile(memFile); err != nil {
					fmt.Fprintf(os.Stderr, "could not write memory profile: %v\n", err)
				}
			})

			// trace
			traceFile, err := os.Create(tracePath)
			if err != nil {
				return fmt.Errorf("could not create trace file: %w", err)
			}
			if err := trace.Start(traceFile); err != nil {
				_ = traceFile.Close()
				return fmt.Errorf("could not start trace: %w", err)
			}
			postRunHooks = append(postRunHooks, func() {
				trace.Stop()
				_ = traceFile.Close()
			})

			// Log where results will be written
			log.Info(fmt.Sprintf("profiling enabled: writing profiles to %s", profileDir))
		}

		return nil
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		for _, hook := range postRunHooks {
			hook()
		}
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true
	err := rootCmd.Execute()
	if err != nil {
		var cmd string
		c, _, cerr := rootCmd.Find(os.Args[1:])
		if cerr == nil {
			cmd = c.Name()
		}

		v := "\n"

		if cmd != "" {
			v += fmt.Sprintf("An error occured running the `%s` command\n", cmd)
		} else {
			v += "An error has occured. "
		}

		v += "The error was:\n\n" + fmt.Sprintf("```%s```", err)
		fmt.Println(ascii.Markdown(v))

		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Debug, "debug-log", "D", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Silence, "quiet", "q", false, "Silence (almost all) logging")
	rootCmd.PersistentFlags().BoolVar(&enableProfiling, "profile", false, "Enable CPU, memory, and trace profiling (writes to profiles/<timestamp>/)")
}
