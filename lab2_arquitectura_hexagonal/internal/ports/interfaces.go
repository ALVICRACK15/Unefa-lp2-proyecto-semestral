package ports

import (
	"context"

	"lab2_arquitectura_hexagonal/internal/domain"
)

type RepositorioTransacciones interface {
	BuscarPorIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.TransaccionRecarga, bool, error)
	Guardar(ctx context.Context, tx *domain.TransaccionRecarga) error
}

type PasarelaPagoBancaria interface {
	Cobrar(ctx context.Context, transactionId, msisdn string, amountCents int64) (string, error)
}

type NotificadorAuditoria interface {
	RegistrarEvento(ctx context.Context, eventName, transactionId, idempotencyKey, statusDetail string) error
}
