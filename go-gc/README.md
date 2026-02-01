# GOMEMLIMIT Impact on etcd Client Performance

## Test Setup
- **Data**: 1000 keys, 10KB each (~10MB total)
- **Memory Pressure**: around 430MB ballast allocation
- **Go GC**: GOGC=off (disabled percentage-based GC)

## Results

### Test 1: GOMEMLIMIT=1000MiB
```
GOGC=off GOMEMLIMIT=1000MiB GODEBUG=gctrace=1 go run main.go
```

**Request Latency**: 32.863241ms

**GC Activity**: None during request (sufficient memory headroom)

---

### Test 2: GOMEMLIMIT=500MiB
```
GOGC=off GOMEMLIMIT=500MiB GODEBUG=gctrace=1 go run main.go
```

**Request Latency**: 1.77412281s (54x slower)

**GC Activity**: 
```
gc 1 @0.932s 3%: 0.079+1143+0.040 ms clock, 1.2+0.35/1144/0.050+0.64 ms cpu, 452->479->469 MB, 452 MB goal
```
- GC triggered during request
- 1143ms spent in GC mark phase
- Memory pressure: 452MB → 479MB → 469MB

## Conclusion

GOMEMLIMIT is a protection mechanism that prevents out-of-memory crashes with best effort, but sacrifices application performance when memory pressure is high:

- **54x latency increase** (32ms → 1.77s) with GOMEMLIMIT=500MiB
- GC pause accounts for ~1.1s of the total 1.77s request time
- The Go runtime aggressively triggers GC to stay within the memory limit, trading CPU time for memory safety
- etcd client performance degrades significantly in memory-constrained environments
