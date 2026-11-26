#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static uint64_t protocolVersionData = 5ULL;

struct CDR_Desalineado {
  char tipo_red;
  uint64_t imsi_abonado;
  char bandera_roaming;
  uint32_t celda_id;
  char codigo_estado;
  uint16_t puerto_origen;
};

struct CDR_Optimizado {
  uint64_t imsi_abonado;
  uint32_t celda_id;
  uint16_t puerto_origen;
  char tipo_red;
  char bandera_roaming;
  char codigo_estado;
};

static void printMemorySegments(const int stackSampleCounter, const uint64_t *heapBuffer) {
  printf("Direccion Global (.data): %p\n", (const void *)&protocolVersionData);
  printf("Direccion Local (Stack):  %p\n", (const void *)&stackSampleCounter);
  printf("Direccion Dinamica (Heap):%p\n", (const void *)heapBuffer);
}

static void printUnalignedOffsets(void) {
  printf("\n--- struct CDR_Desalineado (sizeof = %zu bytes) ---\n", sizeof(struct CDR_Desalineado));
  printf("offsetof(tipo_red):        %zu\n", offsetof(struct CDR_Desalineado, tipo_red));
  printf("offsetof(imsi_abonado):    %zu\n", offsetof(struct CDR_Desalineado, imsi_abonado));
  printf("offsetof(bandera_roaming): %zu\n", offsetof(struct CDR_Desalineado, bandera_roaming));
  printf("offsetof(celda_id):        %zu\n", offsetof(struct CDR_Desalineado, celda_id));
  printf("offsetof(codigo_estado):   %zu\n", offsetof(struct CDR_Desalineado, codigo_estado));
  printf("offsetof(puerto_origen):   %zu\n", offsetof(struct CDR_Desalineado, puerto_origen));
}

static void printOptimizedOffsets(void) {
  printf("\n--- struct CDR_Optimizado (sizeof = %zu bytes) ---\n", sizeof(struct CDR_Optimizado));
  printf("offsetof(imsi_abonado):    %zu\n", offsetof(struct CDR_Optimizado, imsi_abonado));
  printf("offsetof(celda_id):        %zu\n", offsetof(struct CDR_Optimizado, celda_id));
  printf("offsetof(puerto_origen):   %zu\n", offsetof(struct CDR_Optimizado, puerto_origen));
  printf("offsetof(tipo_red):        %zu\n", offsetof(struct CDR_Optimizado, tipo_red));
  printf("offsetof(bandera_roaming): %zu\n", offsetof(struct CDR_Optimizado, bandera_roaming));
  printf("offsetof(codigo_estado):   %zu\n", offsetof(struct CDR_Optimizado, codigo_estado));
}

static void printCacheAndBatchSavings(size_t unalignedBytes, size_t optimizedBytes) {
  const size_t l1CacheLineSize = 64;
  const double tenMillionCount = 10000000.0;
  const double oneMillionDivisor = 1000000.0;

  size_t unalignedPerCacheLine = l1CacheLineSize / unalignedBytes;
  size_t optimizedPerCacheLine = l1CacheLineSize / optimizedBytes;

  double unalignedTotalRam = tenMillionCount * (double)unalignedBytes;
  double optimizedTotalRam = tenMillionCount * (double)optimizedBytes;
  double ramSavedBytes = unalignedTotalRam - optimizedTotalRam;
  double ramSavedMillions = ramSavedBytes / oneMillionDivisor;

  printf("\n--- Cache L1 (64B) y Ahorro en 10,000,000 de Registros CDR ---\n");
  printf("CDR_Desalineado por linea de 64B: %zu\n", unalignedPerCacheLine);
  printf("CDR_Optimizado por linea de 64B:  %zu\n", optimizedPerCacheLine);
  printf("RAM CDR_Desalineado (10M):        %.0f bytes (%.2f millones de bytes)\n",
         unalignedTotalRam, unalignedTotalRam / oneMillionDivisor);
  printf("RAM CDR_Optimizado (10M):         %.0f bytes (%.2f millones de bytes)\n",
         optimizedTotalRam, optimizedTotalRam / oneMillionDivisor);
  printf("Ahorro de RAM (10M):              %.0f bytes (%.2f millones de bytes)\n",
         ramSavedBytes, ramSavedMillions);
}

int main(void) {
  int stackSampleCounter = 42;
  uint64_t *heapBuffer = (uint64_t *)malloc(sizeof(uint64_t) * 4);

  if (heapBuffer == NULL) {
    return 1;
  }

  heapBuffer[0] = protocolVersionData;
  heapBuffer[1] = 100ULL;
  heapBuffer[2] = 200ULL;
  heapBuffer[3] = 300ULL;

  printMemorySegments(stackSampleCounter, heapBuffer);
  printUnalignedOffsets();
  printOptimizedOffsets();
  printCacheAndBatchSavings(sizeof(struct CDR_Desalineado), sizeof(struct CDR_Optimizado));

  free(heapBuffer);
  heapBuffer = NULL;

  return 0;
}
