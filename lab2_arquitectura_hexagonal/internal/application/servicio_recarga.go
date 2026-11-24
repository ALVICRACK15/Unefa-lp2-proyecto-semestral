package application

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"lab2_arquitectura_hexagonal/internal/domain"
	"lab2_arquitectura_hexagonal/internal/ports"
)

var (
	ErrServicioNoInicializado = errors.New("el servicio de recarga no tiene todas sus dependencias")
	ErrPagoRechazado          = errors.New("la pasarela bancaria rechazo el cobro")
)

type ComandoRecarga struct {
	idempotencyKey string
	msisdn         string
	amountCents    int64
}

func NuevoComandoRecarga(idempotencyKey, msisdn string, amountCents int64) ComandoRecarga {
	return ComandoRecarga{
		idempotencyKey: idempotencyKey,
		msisdn:         msisdn,
		amountCents:    amountCents,
	}
}

func (c ComandoRecarga) IdempotencyKey() string {
	return c.idempotencyKey
}

func (c ComandoRecarga) MSISDN() string {
	return c.msisdn
}

func (c ComandoRecarga) AmountCents() int64 {
	return c.amountCents
}

type ServicioRecarga struct {
	repo        ports.RepositorioTransacciones
	pasarela    ports.PasarelaPagoBancaria
	auditoria   ports.NotificadorAuditoria
	txCounter   atomic.Uint64
}

func NuevoServicioRecarga(
	repo ports.RepositorioTransacciones,
	pasarela ports.PasarelaPagoBancaria,
	auditoria ports.NotificadorAuditoria,
) (*ServicioRecarga, error) {
	if repo == nil || pasarela == nil || auditoria == nil {
		return nil, ErrServicioNoInicializado
	}
	return &ServicioRecarga{
		repo:      repo,
		pasarela:  pasarela,
		auditoria: auditoria,
	}, nil
}

func (s *ServicioRecarga) EjecutarRecarga(ctx context.Context, cmd ComandoRecarga) (*domain.TransaccionRecarga, error) {
	existente, found, err := s.repo.BuscarPorIdempotencyKey(ctx, cmd.IdempotencyKey())
	if err != nil {
		return nil, err
	}
	if found {
		_ = s.auditoria.RegistrarEvento(ctx, "RECARGA_IDEMPOTENTE_RETORNADA", existente.ID(), cmd.IdempotencyKey(), string(existente.Estado()))
		return existente, nil
	}

	nextSeq := s.txCounter.Add(1)
	generatedTransactionId := fmt.Sprintf("tx-%06d", nextSeq)

	tx, err := domain.NuevaTransaccionRecarga(generatedTransactionId, cmd.IdempotencyKey(), cmd.MSISDN(), cmd.AmountCents())
	if err != nil {
		_ = s.auditoria.RegistrarEvento(ctx, "RECARGA_VALIDACION_FALLIDA", generatedTransactionId, cmd.IdempotencyKey(), err.Error())
		return nil, err
	}

	if err := s.repo.Guardar(ctx, tx); err != nil {
		return nil, err
	}

	if _, err := s.pasarela.Cobrar(ctx, tx.ID(), tx.MSISDN(), tx.MontoCentavos()); err != nil {
		_ = tx.Rechazar(err.Error())
		_ = s.repo.Guardar(ctx, tx)
		_ = s.auditoria.RegistrarEvento(ctx, "RECARGA_RECHAZADA", tx.ID(), tx.IdempKey(), tx.MotivoRechazo())
		return tx, fmt.Errorf("%w: %v", ErrPagoRechazado, err)
	}

	if err := tx.Aprobar(); err != nil {
		return nil, err
	}

	if err := s.repo.Guardar(ctx, tx); err != nil {
		return nil, err
	}

	_ = s.auditoria.RegistrarEvento(ctx, "RECARGA_APROBADA", tx.ID(), tx.IdempKey(), string(tx.Estado()))
	return tx, nil
}
