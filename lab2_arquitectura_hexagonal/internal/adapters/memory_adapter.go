package adapters

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"lab2_arquitectura_hexagonal/internal/application"
	"lab2_arquitectura_hexagonal/internal/domain"
	"lab2_arquitectura_hexagonal/internal/ports"
)

var ErrMontoInvalidoPasarela = errors.New("monto invalido en pasarela bancaria")

type RepositorioMemoria struct {
	mu          sync.RWMutex
	porKey      map[string]*domain.TransaccionRecarga
	porId       map[string]*domain.TransaccionRecarga
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		porKey: make(map[string]*domain.TransaccionRecarga),
		porId:  make(map[string]*domain.TransaccionRecarga),
	}
}

func (r *RepositorioMemoria) BuscarPorIdempotencyKey(_ context.Context, idempotencyKey string) (*domain.TransaccionRecarga, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tx, ok := r.porKey[idempotencyKey]
	if !ok {
		return nil, false, nil
	}
	return tx, true, nil
}

func (r *RepositorioMemoria) Guardar(_ context.Context, tx *domain.TransaccionRecarga) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.porKey[tx.IdempKey()] = tx
	r.porId[tx.ID()] = tx
	return nil
}

type PasarelaMemoria struct {
	cobrosEjecutados atomic.Int64
	totalCobradoCents atomic.Int64
}

func NuevaPasarelaMemoria() *PasarelaMemoria {
	return &PasarelaMemoria{}
}

func (p *PasarelaMemoria) Cobrar(_ context.Context, transactionId, _ string, amountCents int64) (string, error) {
	if amountCents <= 0 {
		return "", ErrMontoInvalidoPasarela
	}
	p.cobrosEjecutados.Add(1)
	p.totalCobradoCents.Add(amountCents)
	return "auth-" + transactionId, nil
}

func (p *PasarelaMemoria) CobrosEjecutados() int64 {
	return p.cobrosEjecutados.Load()
}

func (p *PasarelaMemoria) TotalCobradoCents() int64 {
	return p.totalCobradoCents.Load()
}

type RegistroAuditoria struct {
	EventName      string
	TransactionId  string
	IdempotencyKey string
	StatusDetail   string
}

type AuditoriaMemoria struct {
	mu      sync.Mutex
	eventos []RegistroAuditoria
}

func NuevaAuditoriaMemoria() *AuditoriaMemoria {
	return &AuditoriaMemoria{
		eventos: make([]RegistroAuditoria, 0),
	}
}

func (a *AuditoriaMemoria) RegistrarEvento(_ context.Context, eventName, transactionId, idempotencyKey, statusDetail string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.eventos = append(a.eventos, RegistroAuditoria{
		EventName:      eventName,
		TransactionId:  transactionId,
		IdempotencyKey: idempotencyKey,
		StatusDetail:   statusDetail,
	})
	return nil
}

func (a *AuditoriaMemoria) TotalEventos() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	return len(a.eventos)
}

type ControladorRecargaPrimario struct {
	servicio *application.ServicioRecarga
}

func NuevoControladorRecargaPrimario(servicio *application.ServicioRecarga) *ControladorRecargaPrimario {
	return &ControladorRecargaPrimario{
		servicio: servicio,
	}
}

func (c *ControladorRecargaPrimario) ProcesarRecarga(ctx context.Context, idempotencyKey, msisdn string, amountCents int64) (*domain.TransaccionRecarga, error) {
	cmd := application.NuevoComandoRecarga(idempotencyKey, msisdn, amountCents)
	return c.servicio.EjecutarRecarga(ctx, cmd)
}

var (
	_ ports.RepositorioTransacciones = (*RepositorioMemoria)(nil)
	_ ports.PasarelaPagoBancaria     = (*PasarelaMemoria)(nil)
	_ ports.NotificadorAuditoria     = (*AuditoriaMemoria)(nil)
)
