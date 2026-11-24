package domain

import (
	"errors"
	"regexp"
	"strings"
)

type EstadoTransaccion string

const (
	EstadoPendiente EstadoTransaccion = "PENDIENTE"
	EstadoAprobada  EstadoTransaccion = "APROBADA"
	EstadoRechazada EstadoTransaccion = "RECHAZADA"
)

var (
	ErrEmptyTransactionID      = errors.New("el identificador de transaccion es obligatorio")
	ErrEmptyIdempotencyKey     = errors.New("la llave de idempotencia es obligatoria")
	ErrInvalidMSISDN           = errors.New("el msisdn no cumple el formato internacional e164")
	ErrInvalidAmountCents      = errors.New("el monto en centavos debe ser estrictamente positivo")
	ErrInvalidStateTransition  = errors.New("transicion de estado invalida")
	ErrRejectionReasonRequired = errors.New("el motivo de rechazo es obligatorio")

	e164Regex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
)

type TransaccionRecarga struct {
	transactionId   string
	idempotencyKey  string
	msisdn          string
	amountCents     int64
	status          EstadoTransaccion
	rejectionReason string
}

func NuevaTransaccionRecarga(id, idempKey, msisdn string, montoCentavos int64) (*TransaccionRecarga, error) {
	cleanTransactionId := strings.TrimSpace(id)
	if cleanTransactionId == "" {
		return nil, ErrEmptyTransactionID
	}

	cleanIdempotencyKey := strings.TrimSpace(idempKey)
	if cleanIdempotencyKey == "" {
		return nil, ErrEmptyIdempotencyKey
	}

	cleanMsisdn := strings.TrimSpace(msisdn)
	if !e164Regex.MatchString(cleanMsisdn) {
		return nil, ErrInvalidMSISDN
	}

	if montoCentavos <= 0 {
		return nil, ErrInvalidAmountCents
	}

	return &TransaccionRecarga{
		transactionId:  cleanTransactionId,
		idempotencyKey: cleanIdempotencyKey,
		msisdn:         cleanMsisdn,
		amountCents:    montoCentavos,
		status:         EstadoPendiente,
	}, nil
}

func (t *TransaccionRecarga) ID() string {
	return t.transactionId
}

func (t *TransaccionRecarga) IdempKey() string {
	return t.idempotencyKey
}

func (t *TransaccionRecarga) MSISDN() string {
	return t.msisdn
}

func (t *TransaccionRecarga) MontoCentavos() int64 {
	return t.amountCents
}

func (t *TransaccionRecarga) Estado() EstadoTransaccion {
	return t.status
}

func (t *TransaccionRecarga) MotivoRechazo() string {
	return t.rejectionReason
}

func (t *TransaccionRecarga) Aprobar() error {
	if t.status != EstadoPendiente {
		return ErrInvalidStateTransition
	}
	t.status = EstadoAprobada
	t.rejectionReason = ""
	return nil
}

func (t *TransaccionRecarga) Rechazar(motivo string) error {
	if t.status != EstadoPendiente {
		return ErrInvalidStateTransition
	}
	cleanReason := strings.TrimSpace(motivo)
	if cleanReason == "" {
		return ErrRejectionReasonRequired
	}
	t.status = EstadoRechazada
	t.rejectionReason = cleanReason
	return nil
}
