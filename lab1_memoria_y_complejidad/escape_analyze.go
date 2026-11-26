package main

import (
	"fmt"
	"time"
)

type Sesion5G struct {
	sessionId    uint64
	subscriberId int
	cellCode     int
	activeFlag   bool
}

func CrearEnStackPorValor(sessionId uint64, subscriberId int, cellCode int) Sesion5G {
	sessionSnapshot := Sesion5G{
		sessionId:    sessionId,
		subscriberId: subscriberId,
		cellCode:     cellCode,
		activeFlag:   true,
	}
	return sessionSnapshot
}

func CrearConEscapeAlHeap(sessionId uint64, subscriberId int, cellCode int) *Sesion5G {
	sessionSnapshot := Sesion5G{
		sessionId:    sessionId,
		subscriberId: subscriberId,
		cellCode:     cellCode,
		activeFlag:   true,
	}
	return &sessionSnapshot
}

func countMatchesLinear(subscribers []int, targetIds []int) int {
	matchedTotal := 0
	for _, targetId := range targetIds {
		for _, candidateId := range subscribers {
			if candidateId == targetId {
				matchedTotal++
				break
			}
		}
	}
	return matchedTotal
}

func countMatchesHashMap(indexedSubscribers map[int]struct{}, targetIds []int) int {
	matchedTotal := 0
	for _, targetId := range targetIds {
		if _, found := indexedSubscribers[targetId]; found {
			matchedTotal++
		}
	}
	return matchedTotal
}

func measureSearchStrategies(datasetSize int, querySampleSize int) {
	subscribers := make([]int, datasetSize)
	indexedSubscribers := make(map[int]struct{}, datasetSize)

	for idx := 0; idx < datasetSize; idx++ {
		subscriberKey := idx*10 + 1
		subscribers[idx] = subscriberKey
		indexedSubscribers[subscriberKey] = struct{}{}
	}

	targetIds := make([]int, querySampleSize)
	for qIdx := 0; qIdx < querySampleSize; qIdx++ {
		sampleOffset := (qIdx * 997) % datasetSize
		targetIds[qIdx] = subscribers[sampleOffset]
	}

	linearStartTime := time.Now()
	linearHits := countMatchesLinear(subscribers, targetIds)
	linearElapsedUs := time.Since(linearStartTime).Microseconds()

	hashStartTime := time.Now()
	hashHits := countMatchesHashMap(indexedSubscribers, targetIds)
	hashElapsedUs := time.Since(hashStartTime).Microseconds()

	fmt.Printf("N=%d, Consultas=%d | Lineal O(n): %d us (%d aciertos) | Mapa O(1): %d us (%d aciertos)\n",
		datasetSize, querySampleSize, linearElapsedUs, linearHits, hashElapsedUs, hashHits)
}

func main() {
	stackSession := CrearEnStackPorValor(9001, 1001, 501)
	heapSession := CrearConEscapeAlHeap(9002, 1002, 502)

	fmt.Printf("Sesion Stack: %+v | Sesion Heap: %+v\n", stackSession, *heapSession)
	measureSearchStrategies(100000, 1000)
}
