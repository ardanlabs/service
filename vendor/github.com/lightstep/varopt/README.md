[![Docs](https://godoc.org/github.com/lightstep/varopt?status.svg)](https://godoc.org/github.com/lightstep/varopt)

# VarOpt Sampling Algorithm

This is an implementation of VarOpt, an unbiased weighted sampling
algorithm described in the paper [Stream sampling for variance-optimal
estimation of subset sums](https://arxiv.org/pdf/0803.0473.pdf) (2008)
by Edith Cohen, Nick Duffield, Haim Kaplan, Carsten Lund, and Mikkel
Thorup.

VarOpt is a reservoir-type sampler that maintains a fixed-size sample
and provides a mechanism for merging unequal-weight samples.

This repository also includes a simple reservoir sampling algorithm,
often useful in conjunction with weighed reservoir sampling, that
implements Algorithm R from [Random sampling with a
reservoir](https://en.wikipedia.org/wiki/Reservoir_sampling#Algorithm_R)
(1985) by Jeffrey Vitter.

## Usage: Natural Weights

A typical use of VarOpt sampling is to estimate network flows using
sample packets.  In this use-case, the weight applied to each sample
is the size of the packet.  Because VarOpt computes an unbiased
sample, sample data points can be summarized along secondary
dimensions.  For example, we can select a subset of sampled packets
according to a secondary attribute, sum the sample weights, and the
result is expected to equal the size of packets corresponding to the
secondary attribute from the original population.

See [weighted_test.go](https://github.com/lightstep/varopt/blob/master/weighted_test.go) for an example.

## Usage: Inverse-probability Weights

Another use for VarOpt sampling uses inverse-probability weights to
estimate frequencies while simultaneously controlling sample
diversity.  Suppose a sequence of observations can be naturally
categorized into N different buckets.  The goal in this case is to
compute a sample where each bucket is well represented, while
maintaining frequency estimates.

In this use-case, the weight assigned to each observation is the
inverse probability of the bucket it belongs to.  The result of
weighted sampling with inverse-probability weights is a uniform
expectated value; in this example we expect an equal number of
observations falling into each bucket.  Each observation represents a
frequency of its sample weight (computed by VarOpt) divided by its
original weight (the inverse-probability).

See [frequency_test.go](https://github.com/lightstep/varopt/blob/master/frequency_test.go) for an example.

## Usage: Merging Samples

VarOpt supports merging independently collected samples one
observation at a time.  This is useful for building distributed
sampling schemes.  In this use-case, each node in a distributed system
computes a weighted sample.  To combine samples, simply input all the
observations and their corresponding weights into a new VarOpt sample.