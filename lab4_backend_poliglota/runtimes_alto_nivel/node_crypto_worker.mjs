import { createHash, pbkdf2Sync } from "node:crypto";
import { Worker, isMainThread, parentPort, workerData } from "node:worker_threads";
import { fileURLToPath } from "node:url";

function computeHashPayload(rawSecret, saltText, iterationsCount) {
  const derivedKey = pbkdf2Sync(rawSecret, saltText, iterationsCount, 32, "sha256");
  const digestHex = createHash("sha256").update(derivedKey).digest("hex");
  return digestHex;
}

function runCryptoOffloadWorker(payloadData) {
  const currentFileUrl = fileURLToPath(import.meta.url);
  return new Promise((resolve, reject) => {
    const workerInstance = new Worker(currentFileUrl, {
      workerData: payloadData,
    });

    workerInstance.on("message", (workerResult) => {
      resolve(workerResult);
    });

    workerInstance.on("error", (workerError) => {
      reject(workerError);
    });

    workerInstance.on("exit", (exitCode) => {
      if (exitCode !== 0) {
        reject(new Error(`Worker finalizado con codigo ${exitCode}`));
      }
    });
  });
}

if (!isMainThread) {
  const { secretToken, saltValue, iterations } = workerData;
  const computedDigest = computeHashPayload(secretToken, saltValue, iterations);
  parentPort.postMessage({ digestHex: computedDigest, iterations });
} else {
  const heartbeatTimer = setInterval(() => {
    process.stdout.write(".");
  }, 25);

  const workerResult = await runCryptoOffloadWorker({
    secretToken: "tx-payload-5g-topup-994812",
    saltValue: "telco-fintech-salt-2026",
    iterations: 120000,
  });

  clearInterval(heartbeatTimer);
  console.log(`\nHash completado sin bloquear Event Loop: ${workerResult.digestHex.slice(0, 16)}... (iteraciones=${workerResult.iterations})`);
}
