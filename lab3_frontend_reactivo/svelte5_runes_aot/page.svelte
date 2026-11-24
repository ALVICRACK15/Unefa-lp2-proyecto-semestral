<script lang="ts">
  interface CellTelemetryDTO {
    cellId: string;
    latencyMs: number;
    packetLossPct: number;
    throughputMbps: number;
  }

  interface TelemetryPanelMetrics {
    averageLatencyMs: number;
    criticalCellCount: number;
    totalPackets: number;
  }

  const maxWindowSize = 20;
  const criticalLatencyThresholdMs = 50;

  let packets = $state<CellTelemetryDTO[]>([
    { cellId: "cell-5g-01", latencyMs: 18, packetLossPct: 0.12, throughputMbps: 940 },
    { cellId: "cell-5g-02", latencyMs: 64, packetLossPct: 1.45, throughputMbps: 410 },
    { cellId: "cell-5g-03", latencyMs: 27, packetLossPct: 0.25, throughputMbps: 820 },
    { cellId: "cell-5g-04", latencyMs: 88, packetLossPct: 2.1, throughputMbps: 295 },
  ]);

  let burstSequence = $state<number>(0);
  let lastBurstDurationMs = $state<number>(0);

  let metrics = $derived.by<TelemetryPanelMetrics>(() => {
    if (packets.length === 0) {
      return {
        averageLatencyMs: 0,
        criticalCellCount: 0,
        totalPackets: 0,
      };
    }

    let latencySum = 0;
    let criticalCount = 0;

    for (const item of packets) {
      latencySum += item.latencyMs;
      if (item.latencyMs > criticalLatencyThresholdMs) {
        criticalCount += 1;
      }
    }

    return {
      averageLatencyMs: Number((latencySum / packets.length).toFixed(2)),
      criticalCellCount: criticalCount,
      totalPackets: packets.length,
    };
  });

  function buildBurstPackets(batchSequence: number, burstCount: number): CellTelemetryDTO[] {
    const generatedPackets: CellTelemetryDTO[] = [];

    for (let index = 0; index < burstCount; index += 1) {
      const isCriticalSpike = (index + batchSequence) % 2 === 0;
      const computedLatency = isCriticalSpike ? 58 + ((index * 7) % 45) : 16 + ((index * 5) % 28);
      const computedLoss = Number((((index + batchSequence) % 10) * 0.35).toFixed(2));
      const computedThroughput = isCriticalSpike ? 320 + ((index * 30) % 250) : 780 + ((index * 25) % 220);

      generatedPackets.push({
        cellId: `cell-5g-${((batchSequence * burstCount + index) % 12) + 1}`,
        latencyMs: computedLatency,
        packetLossPct: computedLoss,
        throughputMbps: computedThroughput,
      });
    }

    return generatedPackets;
  }

  function handleInjectBurst() {
    const startPerfMs = performance.now();
    const nextSequence = burstSequence + 1;
    const burstBatch = buildBurstPackets(nextSequence, 6);
    const mergedPackets = [...packets, ...burstBatch];

    packets = mergedPackets.slice(-maxWindowSize);
    burstSequence = nextSequence;

    const endPerfMs = performance.now();
    lastBurstDurationMs = Number((endPerfMs - startPerfMs).toFixed(4));
  }
</script>

<section aria-labelledby="telemetrySvelteHeading">
  <header>
    <h2 id="telemetrySvelteHeading">Telemetria 5G en Tiempo Real (Svelte 5 Runes AOT)</h2>
    <button type="button" onclick={handleInjectBurst}>
      Inyectar Rafaga 5G
    </button>
  </header>

  <div role="region" aria-label="Resumen de Metricas">
    <p>
      Latencia Promedio: <strong>{metrics.averageLatencyMs} ms</strong>
    </p>
    <p>
      Celdas Criticas (&gt; 50 ms): <strong>{metrics.criticalCellCount}</strong>
    </p>
    <p>
      Paquetes en Ventana: <strong>{metrics.totalPackets}</strong>
    </p>
    <output aria-live="polite">
      Ultima Rafaga: {lastBurstDurationMs} ms (Lote #{burstSequence})
    </output>
  </div>

  <div>
    {#each packets as packet, idx (idx)}
      <article>
        <h3>{packet.cellId}</h3>
        <p>Latencia: {packet.latencyMs} ms</p>
        <p>Perdida de Paquetes: {packet.packetLossPct}%</p>
        <p>Throughput: {packet.throughputMbps} Mbps</p>
      </article>
    {/each}
  </div>
</section>
