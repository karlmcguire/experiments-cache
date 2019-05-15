Experimenting with cache implementations for [ristretto](https://github.com/dgraph-io/ristretto).

## Requirements

### 1. Concurrent

### 2. Memory-bounded

* Limit cache to configurable max memory usage (see [runtime][] package)

[runtime]: https://golang.org/pkg/runtime/

### 3. Scale well as number of cores and goroutines increase

* Compete with Caffeine [performance][]

[performance]: https://github.com/ben-manes/caffeine/wiki/Benchmarks

### 4. Scale well under non-random key access distribution (e.g. Zipf)

* Create Zipf + trace benchmarking library (see [stress][] for example)

[stress]: https://github.com/xba/stress

### 5. High cache hit ratio

* Try to be equal to or better than Caffeine [efficiency][]

[efficiency]: https://github.com/ben-manes/caffeine/wiki/Efficiency

### 6. Minimize Go garbage collection

* BigCache, FreeCache, etc. do this already

## Approaches

### Failed

#### Go map with sync.Mutex (2, 3, 4)

* Doesn't keep track of current memory usage (2)
* Extreme contention (3, 4)

#### Go map with [lock striping][striping] (2, 4)

* Go is slow to release memory to the OS but fast to request it - allocations occurred before release leading to OOM crashing (2)
* Because of Zipf's law a few shards will have high contention (4)

[striping]: https://netjs.blogspot.com/2016/05/lock-striping-in-java-concurrency.html

#### LRU cache ([groupcache][]) (2, 3, 4)

* Hard to estimate complex data structure size (2)
* Every read is a write moving an element in a linked list, causing severe contention (3, 4)
* Despite efforts around lazy eviction, still had severe contention (3, 4)

[groupcache]: https://github.com/golang/groupcache/tree/master/lru

#### Striped LRU cache (4)

* From experience with striped map shards, this would also suffer severe contention (4)
    
### Popular

[Here][comparison] is a comparison of various Go caching libraries and their performance. 

**NOTE**: The above benchmarks are not using a Zipf distribution. Also, I'm pretty sure none of the implementations listed below use a Zipf distribution in their README benchmarks, as the results may not be pretty.

[comparison]: https://github.com/Xeoncross/go-cache-benchmark

#### [BigCache][]

* Uses `map[uint64]uint32` where keys are hashed and values are offsets of entries
* Entries are kept in `[]byte` (reducing GC)
* Can allocate additional memory for new entries when full (FreeCache overwrites entries when full as of 7 months ago)

[BigCache]: https://github.com/allegro/bigcache

#### [FreeCache][]

* Reduces GC by reducing the number of pointers (preallocating memory)
* Always 512 pointers (to buckets/slabs)
    * Data sharded into 256 segments, each segment with 2 pointers:
        * One is the ring buffer storing key-value pairs
        * The other is the "index slice" used to lookup entries
    * Each segment has its own lock

[FreeCache]: https://github.com/coocood/freecache

#### [FastCache][]

* Uses ideas from BigCache:
    * Multiple buckets, each with its own lock (see map lock striping)
    * Each bucket has a `hash(key) -> (key, value) position` map
        * `map[uint64]uint64` for example
    * Each bucket "chunk" is a 64kb `[]byte` (associated with the bucket map values)
* Chunks are allocated off-heap if possible

[FastCache]: https://github.com/VictoriaMetrics/fastcache

## Benchmarking

* Use [math/rand][]'s Zipf generator for mocking cache keys when benchmarking
    * Simulate [real-world][] access distributions
    * Stress test under high contention
* Use same computing environment across all benchmarks

[math/rand]: https://golang.org/pkg/math/rand/#Zipf
[real-world]: https://en.wikipedia.org/wiki/Wikipedia:Does_Wikipedia_traffic_obey_Zipf%27s_law%3F

## Links

* http://highscalability.com/blog/2016/1/25/design-of-a-modern-cache.html
* http://highscalability.com/blog/2019/2/25/design-of-a-modern-cachepart-deux.html
* https://github.com/ben-manes/caffeine/wiki/Design
