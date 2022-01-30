# wordle-solver

`wordle-solver` is one of many wordle solvers. This code is in
an abysmal state, and I do not plan on cleaning it.

# Install

```
go get github.com/ammario/wordle-solver
```

# Basic Usage
[![asciicast](https://asciinema.org/a/evCGbBUnraHdsT8fWEQT66dQJ.svg)](https://asciinema.org/a/evCGbBUnraHdsT8fWEQT66dQJ)

# First Guesses
Guesses are ranked by the average number of words they eliminate across
all possible secrets. This algorithm completes in n^2 time, so the
first guesses (most expensive) are provided here to save you from waiting.

## Best
```
tares | d=3.1994427e+07 | 762.420667ms
rates | d=3.1980355e+07 | 805.875375ms
reals | d=3.1962781e+07 | 781.087583ms
aloes | d=3.1946655e+07 | 817.256416ms
tears | d=3.1939206e+07 | 780.76975ms
raise | d=3.1929631e+07 | 781.93075ms
saner | d=3.1923544e+07 | 808.242583ms
earls | d=3.1922125e+07 | 791.569458ms
nares | d=3.1920887e+07 | 782.158083ms
```
## Worst

```
ohhhh | d=1.866956e+07 | 738.506375ms
immix | d=1.8927119e+07 | 764.3465ms
gyppy | d=1.9059331e+07 | 767.566666ms
yukky | d=1.9428584e+07 | 834.419791ms
fuzzy | d=1.9635818e+07 | 818.358875ms
buzzy | d=2.0299159e+07 | 854.904709ms
yummy | d=2.0417051e+07 | 765.501958ms
ahhhh | d=2.0449205e+07 | 748.472125ms
mummy | d=2.0463702e+07 | 774.73175ms
civic | d=2.0629062e+07 | 747.309458ms
```