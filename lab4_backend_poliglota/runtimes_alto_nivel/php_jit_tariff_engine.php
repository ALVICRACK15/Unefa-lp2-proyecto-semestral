<?php

declare(strict_types=1);

enum TipoRed: string
{
    case LTE_4G = '4G_LTE';
    case NR_5G_SA = '5G_SA';
    case ROAMING_5G = '5G_ROAMING';
}

readonly class TariffQuote
{
    public function __construct(
        public string $tariffCode,
        public TipoRed $networkType,
        public int $chargedCents,
    ) {}
}

readonly class JitTariffEngine
{
    public function __construct(
        private int $ratePerMegabyteCents,
    ) {}

    public function calculateTariff(TipoRed $networkType, int $megabytesConsumed): TariffQuote
    {
        $safeMegabytes = max(0, $megabytesConsumed);
        $baseCents = $safeMegabytes * $this->ratePerMegabyteCents;

        $chargedCents = match ($networkType) {
            TipoRed::LTE_4G => $baseCents,
            TipoRed::NR_5G_SA => (int) round($baseCents * 1.25),
            TipoRed::ROAMING_5G => (int) round($baseCents * 2.0),
        };

        $tariffCode = match ($networkType) {
            TipoRed::LTE_4G => 'TARIFF-4G-STD',
            TipoRed::NR_5G_SA => 'TARIFF-5G-SA',
            TipoRed::ROAMING_5G => 'TARIFF-5G-ROAM',
        };

        return new TariffQuote(
            tariffCode: $tariffCode,
            networkType: $networkType,
            chargedCents: $chargedCents,
        );
    }
}

$tariffEngine = new JitTariffEngine(ratePerMegabyteCents: 10);
$sampleQuotes = [
    $tariffEngine->calculateTariff(TipoRed::LTE_4G, 100),
    $tariffEngine->calculateTariff(TipoRed::NR_5G_SA, 100),
    $tariffEngine->calculateTariff(TipoRed::ROAMING_5G, 50),
];

foreach ($sampleQuotes as $quoteItem) {
    echo sprintf(
        "%s (%s): %d centavos\n",
        $quoteItem->tariffCode,
        $quoteItem->networkType->value,
        $quoteItem->chargedCents
    );
}
