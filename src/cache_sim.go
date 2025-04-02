package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Request struct {
	Tenant          string
	Product         string
	Timestamp       int64
	FirstAppearance string
	Size            int64
}

func readTrace(reqChan chan Request) {
	scanner := bufio.NewScanner(os.Stdin)

	if scanner.Scan() {
		_ = scanner.Text()
	}

	for scanner.Scan() {
		row := strings.Split(scanner.Text(), ",")
		var request Request
		request.Tenant = row[0]
		request.Product = row[1]
		request.Timestamp, _ = strconv.ParseInt(row[2], 10, 64)
		/*
		   If it is true, it get the column "cache_status" that describes which status is the first appearance of an item.
		*/
		if use_fa_fix {
			request.FirstAppearance = row[4]
		}
		//By default, use_fa_fix is false, so row[3] is the Size
		request.Size, _ = strconv.ParseInt(row[3], 10, 64)
		reqChan <- request
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Erro ao ler a entrada:", err)
	}

	close(reqChan)
}

type RequestStatus string

const (
	MISS  RequestStatus = "M"
	STALE RequestStatus = "S"
	HIT   RequestStatus = "H"
)

const (
	ONE_MINUTE          int64 = 60000000
	TWENTY_FIVE_MINUTES int64 = ONE_MINUTE * 25
	FIFTY_MINUTES       int64 = ONE_MINUTE * 50
)

type RequestResponse struct {
	request        Request
	status         RequestStatus
	cache_usage    int64
	cache_capacity int64
	capacity       int64
}

type Pair struct {
	tenant  string
	product string
}

func isExpired(req Request, lastAccess int64) bool {
	diff := req.Timestamp - lastAccess
	return (diff > ttl)
}

var lastPurgeStamp int64 = 0
var lastMaxMinStamp int64 = 1670976000001000

/*
Return a boolean value that represents if it's time to verify and
remove expired cached items.
*/
func shouldPurge(now int64) bool {

	doPurge := ((now - lastPurgeStamp) > purge_check_interval)
	lastPurgeStamp = now
	return doPurge
}

func getCacheForTenant(tenant string) *LRUCache {
	if caches, exists := cache[tenant]; exists {
		return caches
	}
	// Create a new cache for the tenant if it does not exist
	newCache := NewLRUCache(cache_size)
	cache[tenant] = newCache
	return newCache
}

/*
Simulates a cache get request. Verify if an item is in the cache storage,
returning the request status result and the cache usage after the request.
- If the item is not cached, add the item and return Miss status.
- If the item is cached, but is expired, renew item and return Stale status.
- If the item is cached, and is not expired, renew item and return Hit status.

It',s possible to use admission control, passing as argument the flag -use_ac.
This permits to cache items on second Miss status.
*/
func simulateGet(cache *LRUCache, req Request) RequestStatus {
	if shouldPurge(req.Timestamp) {
		//purge()
	}

	if !admissionControl.Admit(req.Product) {
		return MISS
	}
	/*
	   If it is true, it returns the exact status that appears in the first requests of that item.
	*/
	if use_fa_fix {
		if req.FirstAppearance == "H" {
			cache.Add(req.Product, req.Timestamp, req.Size)
			return HIT
		}

		if req.FirstAppearance == "S" {
			cache.Add(req.Product, req.Timestamp, req.Size)
			return STALE
		}
	}

	ts, found := cache.Add(req.Product, req.Timestamp, req.Size)

	if found {
		if isExpired(req, ts.(int64)) {
			return STALE
		}

		return HIT
	} else {

		return MISS
	}
}

/*
	func purge() {
		inactiveItems := list.New()
		//FIXME: i think we are disclosing internal details of this cache implementation
		for e := cache.lruList.Front(); e != nil; e = e.Next() {
			if lastPurgeStamp-e.Value.(*cacheItem).value.(int64) > exp_ttl {
				inactiveItems.PushFront(e.Value.(*cacheItem).key)
			}
		}

		cache.RemoveAll(inactiveItems)
	}
*/
func simulate(trace chan Request, output chan RequestResponse) {
	first_time := 1670976000001000
	for req := range trace {
		interval := (int(req.Timestamp) - first_time) / int(maxmin_time)
		status := MISS
		cache_usage := 0
		cache_capacity := 0
		//no part
		if interval == 0 {
			tenant_cache := getCacheForTenant(req.Tenant)
			tenant_cache.Add(req.Product, req.Timestamp, req.Size)
			status = simulateGet(cache_nopart, req)
			cache_usage = int(cache_nopart.cacheUsage)
			cache_capacity = int(cache_nopart.capacity)
		} else {
			// maxmin
			if req.Timestamp-lastMaxMinStamp >= maxmin_time {
				maxmin(interval)
				lastMaxMinStamp = req.Timestamp
				for tenant := range demands {
					demands[tenant] = 0
				}

			}

			caches := getCacheForTenant(req.Tenant)
			status = simulateGet(caches, req)
			cache_usage = int(cache[req.Tenant].Size())
			cache_capacity = int(cache[req.Tenant].capacity)

		}

		if _, exists := demands[req.Tenant]; exists {
			demands[req.Tenant] += req.Size
		} else {
			demands[req.Tenant] = req.Size
		}

		output <- RequestResponse{
			request:        req,
			status:         status,
			cache_usage:    int64(cache_usage),
			cache_capacity: int64(cache_capacity),
			capacity:       caps[interval][req.Tenant],
		}
	}

	close(output)
}
func removeCapacity(tenant string, capacity int64) {
	for cache[tenant].cacheUsage > capacity {
		lastElem := cache[tenant].lruList.Back()
		cache[tenant].Remove(lastElem.Value.(*cacheItem).key)
	}
	cache[tenant].capacity = capacity
}

func maxmin(interval int) {
	capacities := calculateCapacities(demands)
	caps[interval] = capacities
	for tenant := range cache {
		if _, exists := capacities[tenant]; !exists {
			delete(cache, tenant)
		} else {
			if cache[tenant].capacity < capacities[tenant] {
				cache[tenant].capacity = capacities[tenant]
			} else {
				removeCapacity(tenant, capacities[tenant])
			}
		}
	}
}

func calculateCapacities(demands map[string]int64) map[string]int64 {
	type kv struct {
		key   string
		value int64
	}

	var sorted_slice []kv
	for k, v := range demands {
		sorted_slice = append(sorted_slice, kv{k, v})
	}

	sort.Slice(sorted_slice, func(i, j int) bool {
		return sorted_slice[i].value < sorted_slice[j].value
	})

	var demand_values []int64
	for _, kv := range sorted_slice {
		demand_values = append(demand_values, kv.value)
	}

	capacities := MaxMinFairness(demand_values, cache_size)

	result := make(map[string]int64)
	for i, kv := range sorted_slice {
		result[kv.key] = capacities[i]
	}

	return result
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// change it to any of the algorithms
func MaxMinFairness(demands []int64, capacity int64) []int64 {
	capacityRemaining := capacity
	output := make([]int64, len(demands))

	for i, demand := range demands {
		share := capacityRemaining / int64(len(demands)-i)
		allocation := min(share, demand)

		if i == len(demands)-1 {
			if demand >= capacityRemaining {
				allocation = capacityRemaining
			}
		}

		output[i] = allocation
		capacityRemaining -= allocation
	}

	if capacityRemaining != 0 {
		for i := range output {
			add := capacityRemaining / int64(len(output))
			output[i] = output[i] + add
		}
	}

	return output
}

func toMicroseconds(time int64) int64 {
	return time * ONE_MINUTE
}

var cache map[string]*LRUCache
var demands map[string]int64
var admissionControl *AdmissionControl
var cache_nopart *LRUCache
var caps map[int]map[string]int64

var ttl int64
var exp_ttl int64
var purge_check_interval int64
var cache_size int64
var use_ac bool
var use_fa_fix bool
var maxmin_time int64

/*
Get arguments from command line execution.
*/
func get_arguments() {
	flag.Int64Var(&cache_size, "cache_size", 0, "How much concurrent items the cache can hold")
	flag.Int64Var(&ttl, "ttl", TWENTY_FIVE_MINUTES, "Cached items time-to-live in minutes")
	flag.Int64Var(&exp_ttl, "exp_ttl", FIFTY_MINUTES, "Cached items inactivation time in minutes")
	flag.Int64Var(&purge_check_interval, "purge_check_interval", ONE_MINUTE, "Verification time to purge inactive items in minutes")
	flag.BoolVar(&use_ac, "use_ac", false, "Use admission control")
	flag.BoolVar(&use_fa_fix, "use_fa_fix", false, "Use First Appearance")
	flag.Int64Var(&maxmin_time, "maxmin_time", 0, "Max min time")

	flag.Parse()

	ttl = toMicroseconds(ttl)
	exp_ttl = toMicroseconds(exp_ttl)
	purge_check_interval = toMicroseconds(purge_check_interval)
	maxmin_time = toMicroseconds(maxmin_time)

	var exit = false
	var msg_error = ""

	if cache_size < 0 {
		msg_error = msg_error + "The cache size cannot be smaller than 0.\n"
		exit = true
	}
	if ttl < 0 {
		msg_error = msg_error + "The ttl (time-to-live) cannot be smaller than 0.\n"
		exit = true
	}
	if exp_ttl < 0 {
		msg_error = msg_error + "The exp_ttl cannot be smaller than 0.\n"
		exit = true
	}
	if purge_check_interval < 0 {
		msg_error = msg_error + "The purge_check_interval cannot be smaller than 0.\n"
		exit = true
	}

	if exit {
		fmt.Fprintf(os.Stderr, msg_error)
		os.Exit(1)
	}
}

// TODO: we are copying structs here and there, maybe pass pointers
func main() {
	get_arguments()

	cache = make(map[string]*LRUCache)
	demands = make(map[string]int64)
	caps = make(map[int]map[string]int64)
	cache_nopart = NewLRUCache(cache_size)

	admissionControl = NewAdmissionControl(use_ac)

	reqChan := make(chan Request, 10000)
	responseChan := make(chan RequestResponse, 10000)

	go readTrace(reqChan)
	go simulate(reqChan, responseChan)

	// AV: Is necessary add the size of products? check
	fmt.Printf("timestamp,product,tenant,cache_response,cache_usage,ac_usage,allocation\n")
	for res := range responseChan {
		fmt.Printf(
			"%d,%s,%s,%s,%d,%d,%d\n",
			res.request.Timestamp,
			res.request.Product,
			res.request.Tenant,
			res.status,
			res.cache_usage,
			res.cache_capacity,
			res.capacity,
		)
	}
}
