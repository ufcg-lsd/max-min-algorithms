# MaxMin Allocation Algorithms

This repository contains implementations of various resource allocation algorithms, including Fair Share, Water Filling and Max-Min LSD, written in Go.

# Algorithms

Describing each one of the algorithms proposed.

## [FairShare Algorithm](algorithms/fairshare.go)

The **`FairShare`** algorithm is a simple implementation of *Max-Min Fairness*. It distributes a fixed `capacity` among users based on their `demands`, ensuring that no user receives more than their demand and the allocation is as fair as possible.

## [WaterFilling Algorithm](algorithms/waterfilling.go)

The **`WaterFilling`** algorithm distributes a fixed `capacity` among users in a level-by-level fashion, similar to how water fills containers. At each step, it gives one unit to each user whose demand is not yet fully satisfied, cycling through until either the total capacity is exhausted or all demands are met.

## [Max-Min LSD Algorithm](algorithms/maxmin-lsd.go)

The **`Max-Min LSD`** algorithm implements an approximate max-min fairness strategy. It processes sorted demands and tries to give each user an equal share of the remaining capacity, updating the share dynamically as users are allocated. The name `LSD` comes from the laboratory where the algorithm was made, **`Laboratório de Sistemas Distribuídos - UFCG`**.