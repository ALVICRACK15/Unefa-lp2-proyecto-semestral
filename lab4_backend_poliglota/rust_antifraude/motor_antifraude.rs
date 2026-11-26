use std::sync::{Arc, Mutex};
use std::thread;

#[derive(Debug)]
struct AlertaFraude<'a> {
    id_tx: &'a str,
    msisdn: &'a str,
    motivo: &'a str,
    monto_centavos: u64,
}

#[derive(Debug, Default)]
struct EstadisticasRiesgo {
    totalAudited: u64,
    alertCount: u64,
    flaggedAmountCents: u64,
    recordedAlerts: Vec<(String, u64)>,
}

#[derive(Clone, Copy)]
struct TransactionInput<'a> {
    txId: &'a str,
    subscriberMsisdn: &'a str,
    amountCents: u64,
    riskScore: u16,
}

fn evaluateSingleTransaction<'a>(txItem: TransactionInput<'a>) -> Option<AlertaFraude<'a>> {
    let highAmountThreshold: u64 = 250000;
    let highRiskThreshold: u16 = 75;

    if txItem.amountCents >= highAmountThreshold && txItem.riskScore >= highRiskThreshold {
        return Some(AlertaFraude {
            id_tx: txItem.txId,
            msisdn: txItem.subscriberMsisdn,
            motivo: "MONTO_Y_RIESGO_ATIPICO",
            monto_centavos: txItem.amountCents,
        });
    }

    if txItem.riskScore >= 90 {
        return Some(AlertaFraude {
            id_tx: txItem.txId,
            msisdn: txItem.subscriberMsisdn,
            motivo: "PATRON_ANOMALO_RED",
            monto_centavos: txItem.amountCents,
        });
    }

    None
}

fn auditChunkWorker(
    chunkSlice: &[TransactionInput<'_>],
    sharedStats: Arc<Mutex<EstadisticasRiesgo>>,
) {
    let mut localAudited: u64 = 0;
    let mut localAlerts: u64 = 0;
    let mut localFlaggedCents: u64 = 0;
    let mut localAlertSummaries: Vec<(String, u64)> = Vec::new();

    for txItem in chunkSlice {
        localAudited += 1;
        if let Some(alert) = evaluateSingleTransaction(*txItem) {
            localAlerts += 1;
            localFlaggedCents += alert.monto_centavos;
            localAlertSummaries.push((format!("{}:{}", alert.id_tx, alert.motivo), alert.monto_centavos));
        }
    }

    let mut guard = sharedStats.lock().unwrap_or_else(|poisoned| poisoned.into_inner());
    guard.totalAudited += localAudited;
    guard.alertCount += localAlerts;
    guard.flaggedAmountCents += localFlaggedCents;
    guard.recordedAlerts.extend(localAlertSummaries);
}

fn buildSampleTransactions() -> Vec<TransactionInput<'static>> {
    vec![
        TransactionInput { txId: "TX-1001", subscriberMsisdn: "+584121110001", amountCents: 320000, riskScore: 82 },
        TransactionInput { txId: "TX-1002", subscriberMsisdn: "+584121110002", amountCents: 5000, riskScore: 15 },
        TransactionInput { txId: "TX-1003", subscriberMsisdn: "+584121110003", amountCents: 18000, riskScore: 94 },
        TransactionInput { txId: "TX-1004", subscriberMsisdn: "+584121110004", amountCents: 12000, riskScore: 22 },
        TransactionInput { txId: "TX-1005", subscriberMsisdn: "+584121110005", amountCents: 500000, riskScore: 88 },
        TransactionInput { txId: "TX-1006", subscriberMsisdn: "+584121110006", amountCents: 7500, riskScore: 10 },
        TransactionInput { txId: "TX-1007", subscriberMsisdn: "+584121110007", amountCents: 260000, riskScore: 79 },
        TransactionInput { txId: "TX-1008", subscriberMsisdn: "+584121110008", amountCents: 9000, riskScore: 18 },
    ]
}

fn main() {
    let transactionsBatch = buildSampleTransactions();
    let workerCount: usize = 4;
    let totalItems = transactionsBatch.len();
    let chunkSize = (totalItems + workerCount - 1) / workerCount;
    let sharedStats = Arc::new(Mutex::new(EstadisticasRiesgo::default()));

    thread::scope(|scope| {
        for workerIdx in 0..workerCount {
            let startIdx = workerIdx * chunkSize;
            if startIdx >= totalItems {
                break;
            }
            let endIdx = (startIdx + chunkSize).min(totalItems);
            let sliceRef = &transactionsBatch[startIdx..endIdx];
            let statsClone = Arc::clone(&sharedStats);

            scope.spawn(move || {
                auditChunkWorker(sliceRef, statsClone);
            });
        }
    });

    let finalGuard = sharedStats.lock().unwrap_or_else(|poisoned| poisoned.into_inner());
    println!(
        "Auditadas: {} | Alertas: {} | Monto en Riesgo (centavos): {} | Detalles: {:?}",
        finalGuard.totalAudited,
        finalGuard.alertCount,
        finalGuard.flaggedAmountCents,
        finalGuard.recordedAlerts
    );
}
