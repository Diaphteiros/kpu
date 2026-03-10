## kpu from-seconds

Convert the given amount of seconds into a human-readable duration

### Synopsis

Converts the given amount of seconds into a human-readable duration string.

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
	

```
kpu from-seconds <duration> [flags]
```

### Options

```
  -h, --help   help for from-seconds
```

### SEE ALSO

* [kpu](kpu.md)	 - Improve your k8s cluster interactions

