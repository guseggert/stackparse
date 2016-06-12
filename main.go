package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	util "github.com/whyrusleeping/stackparse/util"
)

func printHelp() {
	fmt.Println("to filter out goroutines from the trace, use the following flags:")
	fmt.Println("--frame-match=FOO")
	fmt.Println("  print only stacks with frames that contain 'FOO'")
	fmt.Println("--frame-not-match=FOO")
	fmt.Println("  print only stacks with no frames containing 'FOO'")
	fmt.Println("--wait-more-than=10m")
	fmt.Println("  print only stacks that have been blocked for more than ten minutes")
	fmt.Println("--wait-less-than=10m")
	fmt.Println("  print only stacks that have been blocked for less than ten minutes")
	fmt.Println("\n")
	fmt.Println("output is by default sorted by waittime ascending, to change this use:")
	fmt.Println("--sort=[stacksize,goronum,waittime]")
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("usage: %s <filter flags> <filename>\n", os.Args[0])
		printHelp()
		return
	}

	var filters []util.Filter
	var compfunc util.StackCompFunc = util.CompWaitTime
	fname := "-"

	// parse flags
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "--") {
			parts := strings.Split(a, "=")
			if len(parts) != 2 {
				fmt.Println("all flags must be --opt=val")
				os.Exit(1)
			}

			switch parts[0] {
			case "--frame-match":
				filters = append(filters, util.HasFrameMatching(parts[1]))
			case "--wait-more-than":
				d, err := time.ParseDuration(parts[1])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				filters = append(filters, util.TimeGreaterThan(d))
			case "--wait-less-than":
				d, err := time.ParseDuration(parts[1])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				filters = append(filters, util.Negate(util.TimeGreaterThan(d)))
			case "--frame-not-match":
				filters = append(filters, util.Negate(util.HasFrameMatching(parts[1])))
			case "--sort":
				switch parts[1] {
				case "goronum":
					compfunc = util.CompGoroNum
				case "stacksize":
					compfunc = util.CompDepth
				case "waittime":
					compfunc = util.CompWaitTime
				default:
					fmt.Println("unknown sorting parameter: ", parts[1])
					os.Exit(1)
				}
			}
		} else {
			fname = a
		}
	}

	var r io.Reader
	if fname == "-" {
		r = os.Stdin
	} else {
		fi, err := os.Open(fname)
		if err != nil {
			panic(err)
		}
		defer fi.Close()

		r = fi
	}

	stacks, err := util.ParseStacks(r)
	if err != nil {
		panic(err)
	}

	sorter := util.StackSorter{
		Stacks:   stacks,
		CompFunc: compfunc,
	}

	sort.Sort(sorter)

	for _, s := range util.ApplyFilters(stacks, filters) {
		s.Print()
	}
}
