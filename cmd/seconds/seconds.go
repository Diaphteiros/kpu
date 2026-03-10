package seconds

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
	"github.com/spf13/cobra"
)

// ToSecondsCmd represents the to-seconds command
var ToSecondsCmd = &cobra.Command{
	Aliases: []string{"to-second", "to-sec"},
	Use:     "to-seconds <duration>",
	Args:    cobra.ExactArgs(1),
	GroupID: cmdgroups.Utilities,
	Short:   "Convert the given duration to seconds",
	Long: `Converts the given duration to seconds.

A duration consists of one or more natural numbers, each followed by a unit, e.g. '1h30m'.

The following units are supported:
- 's' for seconds
- 'm' for minutes
- 'h' for hours
- 'd' for days
- 'w' for weeks (1w = 7d)
- 'M' for months (1M = 30d)
- 'y' for years (1y = 365d)

Note that due to the simplified logic, neither 12 months nor 52 weeks add up to the 365 days of a year.

Examples:

	> kpu to-seconds 1h
	3600

	> kpu to-seconds 1y15d3h
	32842800
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateToSecondsCommand(args)

		d, err := utils.StringToDuration(args[0])
		if err != nil {
			utils.Fatal(1, "error converting duration to seconds: %s", err.Error())
		}

		fmt.Printf("%d\n", int64(d.Seconds()))
	},
}

func ValidateToSecondsCommand(args []string) {
	if len(args) != 1 {
		utils.Fatal(1, "expected exactly one argument, got %d", len(args))
	}
}

// FromSecondsCmd represents the from-seconds command
var FromSecondsCmd = &cobra.Command{
	Aliases: []string{"from-second", "from-sec"},
	Use:     "from-seconds <duration>",
	Args:    cobra.ExactArgs(1),
	GroupID: cmdgroups.Utilities,
	Short:   "Convert the given amount of seconds into a human-readable duration",
	Long: `Converts the given amount of seconds into a human-readable duration string.

The following units are used:
- 's' for seconds
- 'm' for minutes
- 'h' for hours
- 'd' for days
- 'y' for years (1y = 365d)

Examples:

	> kpu from-seconds 3600
	1h

	> kpu from-seconds 32842800
	1y15d3h
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateFromSecondsCommand(args)

		dAsInt64, ok := strconv.ParseInt(strings.TrimSuffix(args[0], "s"), 10, 64)
		if ok != nil {
			utils.Fatal(1, "expected a number, got %s", args[0])
		}
		d := time.Duration(dAsInt64) * time.Second
		fmt.Println(utils.FormatDuration(d))
	},
}

func ValidateFromSecondsCommand(args []string) {
	if len(args) != 1 {
		utils.Fatal(1, "expected exactly one argument, got %d", len(args))
	}
}
