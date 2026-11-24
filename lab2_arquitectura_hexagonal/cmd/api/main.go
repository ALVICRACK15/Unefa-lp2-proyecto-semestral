package main

import (
	"context"
	"fmt"
	"log"

	"lab2_arquitectura_hexagonal/internal/adapters"
	"lab2_arquitectura_hexagonal/internal/application"
)

func main() {
	ctx := context.Background()

	repo := adapters.NuevoRepositorioMemoria()
	pasarela := adapters.NuevaPasarelaMemoria()
	auditoria := adapters.NuevaAuditoriaMemoria()

	servicio, err := application.NuevoServicioRecarga(repo, pasarela, auditoria)
	if err != nil {
		log.Fatal(err)
	}

	controlador := adapters.NuevoControladorRecargaPrimario(servicio)

	txExitosa, err := controlador.ProcesarRecarga(ctx, "550e8400-e29b-41d4-a716-446655440001", "+584141234567", 500)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Caso A (Exitosa): id=%s estado=%s montoCents=%d\n", txExitosa.ID(), txExitosa.Estado(), txExitosa.MontoCentavos())

	_, errInvalido := controlador.ProcesarRecarga(ctx, "550e8400-e29b-41d4-a716-446655440002", "04141234567", 500)
	fmt.Printf("Caso B (E.164 Invalido): err=%v\n", errInvalido)

	retryKey := "550e8400-e29b-41d4-a716-446655440003"
	for intento := 1; intento <= 3; intento++ {
		txRetry, errRetry := controlador.ProcesarRecarga(ctx, retryKey, "+584129876543", 1200)
		fmt.Printf("Caso C (Reintento %d): id=%s estado=%s montoCents=%d err=%v\n", intento, txRetry.ID(), txRetry.Estado(), txRetry.MontoCentavos(), errRetry)
	}

	fmt.Printf("Resumen Pasarela: cobrosEjecutados=%d totalCobradoCents=%d eventosAuditoria=%d\n",
		pasarela.CobrosEjecutados(),
		pasarela.TotalCobradoCents(),
		auditoria.TotalEventos(),
	)
}
