# nin

[![Go Reference](https://pkg.go.dev/badge/github.com/maruel/nin.svg)](https://pkg.go.dev/github.com/maruel/nin)
[![codecov](https://codecov.io/gh/maruel/nin/branch/main/graph/badge.svg?token=KAO6K039PJ)](https://codecov.io/gh/maruel/nin)

Little nin' is [ninja](https://ninja-build.org/)'s little sibling.

Nin is an experimental fork of ninja translated in Go.

## Installation

Install [go1.17](https://go.dev/dl/) or later.

```
go install github.com/maruel/nin/cmd/nin@latest
```

Use `nin` where you would have used `ninja` (or create a symlink).

## Are you serious?

Yeah.

The reason it's possible at all is because
[ninja](https://github.com/ninja-build/ninja) is well written and has a
reasonable amount of unit tests.

## Why?

- The parser can be used as a
  [library](https://pkg.go.dev/github.com/maruel/nin)
  - This opens the door to a lot of opportunity and a real ecosystem
- Refactoring Go >> refactoring C++
  - As I made progress, I saw opportunities for simplification
- Making the code concurrent (e.g. making the parser parallel) is easier
- Plans to have it be stateful and/or do RPCs and change fundamental parts
  - E.g. call directly the RBE backend instead of shelling out reclient?
- It's easier to keep the performance promise in check, and keep it maintainable
  - Go has native benchmarking
  - Go has native CPU and memory profiling
  - Go has native code coverage
  - Go has native [documentation](https://pkg.go.dev/github.com/maruel/nin)
- Since it's GC, and the program runs as a one shot, we can just disable GC and
  save a significant amount of memory management (read: CPU) overhead.

I'll write a better roadmap if the project doesn't crash and burn.

Some people did advent of code 2021, I did a brain teaser instead.

### Concerns

- Go's slice and string bound checking slow things down, I'll have to micro
  optimize a bit.
- Go's function calls are more expensive and the Go compiler inlines less often
  than the C++ compiler. I'll reduce the number of function calls.
- Go's Windows layer is likely slower than raw C++, so I'll probably call raw
  syscall functions on Windows.
- Staying up to date changes done upstream, especially to the file format and
  correctness checks.

## Current state

- Manifest (build.ninja) parsing: 5% faster on average! 📉
- Latency: nin is in the same ballpark (-5%) for building ninja itself.
- CPU usage: about 15% higher, has to be optimized.
- 16 test cases out of 394 (5%) have to be fixed. `git grep Skip..TODO | wc -l`
  versus `git grep "^func Test" | wc -l`.
- Closely tracking upstream as-is.
- Flag parsing is not 100% compatible yet.

See [PERF.md](PERF.md) to learn how to measure performance yourself, since I
know myself enough that I will forget to update the stats above and it will get
better over time.

## ninja

Ninja is a small build system with a focus on speed.
https://ninja-build.org/

See [the manual](https://ninja-build.org/manual.html) or
`doc/manual.asciidoc` included in the distribution for background
and more details.
