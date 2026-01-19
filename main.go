// TCRS CLI - Timecard Recording System Command Line Interface
package main

import "github.com/user/tcrs/cmd"

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	cmd.Execute()
}
