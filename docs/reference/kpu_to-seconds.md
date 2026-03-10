## kpu to-seconds

Convert the given duration to seconds

### Synopsis

Converts the given duration to seconds.

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
	

```
kpu to-seconds <duration> [flags]
```

### Options

```
  -h, --help   help for to-seconds
```

### SEE ALSO

* [kpu](kpu.md)	 - Improve your k8s cluster interactions

