package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Transaccion struct {
	transactionId string
	msisdn        string
	amountCents   int64
}

type ResultadoAuditoria struct {
	transactionId string
	approvedFlag  bool
	processedCost int64
}

type ProcesadorTransaccion interface {
	Procesar(tx Transaccion) ResultadoAuditoria
}

type motorEvaluacionReglas struct {
	maxAllowedCents int64
}

func (engine *motorEvaluacionReglas) Procesar(tx Transaccion) ResultadoAuditoria {
	isValid := tx.amountCents > 0 && tx.amountCents <= engine.maxAllowedCents
	return ResultadoAuditoria{
		transactionId: tx.transactionId,
		approvedFlag:  isValid,
		processedCost: tx.amountCents,
	}
}

func EjecutarWorkerPool(
	workerCount int,
	processor ProcesadorTransaccion,
	jobs <-chan Transaccion,
	results chan<- ResultadoAuditoria,
	approvedCounter *atomic.Int64,
	rejectedCounter *atomic.Int64,
) {
	var wg sync.WaitGroup

	for workerIdx := 0; workerIdx < workerCount; workerIdx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for txItem := range jobs {
				auditOutcome := processor.Procesar(txItem)
				if auditOutcome.approvedFlag {
					approvedCounter.Add(1)
				} else {
					rejectedCounter.Add(1)
				}
				results <- auditOutcome
			}
		}()
	}

	wg.Wait()
	close(results)
}

func main() {
	const poolSize = 4
	const totalTransactions = 12

	ruleProcessor := &motorEvaluacionReglas{maxAllowedCents: 200000}
	jobsChannel := make(chan Transaccion, totalTransactions)
	resultsChannel := make(chan ResultadoAuditoria, totalTransactions)
	var approvedTotal atomic.Int64
	var rejectedTotal atomic.Int64

	for idx := 0; idx < totalTransactions; idx++ {
		amountValue := int64((idx + 1) * 25000)
		jobsChannel <- Transaccion{
			transactionId: fmt.Sprintf("GO-TX-%03d", idx+1),
			msisdn:        fmt.Sprintf("+58414555%04d", idx+1),
			amountCents:   amountValue,
		}
	}
	close(jobsChannel)

	EjecutarWorkerPool(
		poolSize,
		ruleProcessor,
		jobsChannel,
		resultsChannel,
		&approvedTotal,
		&rejectedTotal,
	)

	collectedResults := make([]ResultadoAuditoria, 0, totalTransactions)
	for resItem := range resultsChannel {
		collectedResults = append(collectedResults, resItem)
	}

	fmt.Printf("Total: %d | Aprobadas: %d | Rechazadas: %d\n",
		len(collectedResults), approvedTotal.Load(), rejectedTotal.Load())
}
