// Copyright 2019, LightStep Inc.

/*
Package varopt contains an implementation of VarOpt, an unbiased weighted
sampling algorithm described in the paper "Stream sampling for
variance-optimal estimation of subset sums"
https://arxiv.org/pdf/0803.0473.pdf (2008), by Edith Cohen, Nick
Duffield, Haim Kaplan, Carsten Lund, and Mikkel Thorup.

VarOpt is a reservoir-type sampler that maintains a fixed-size sample
and provides a mechanism for merging unequal-weight samples.

This package also includes a simple reservoir sampling algorithm,
often useful in conjunction with weighed reservoir sampling, using
Algorithm R from "Random sampling with a
reservoir", https://en.wikipedia.org/wiki/Reservoir_sampling#Algorithm_R
(1985), by Jeffrey Vitter.

See https://github.com/lightstep/varopt/blob/master/README.md for
more detail.
*/
package varopt
