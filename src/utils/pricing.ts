import type { PricingRule } from '@/types';

export function roundPrice(value: number, mode: PricingRule['roundingMode']) {
  if (mode === 'whole') {
    return Math.ceil(value);
  }

  if (mode === 'x.9') {
    const candidate = Math.floor(value) + 0.9;
    return candidate >= value ? candidate : Math.ceil(value) + 0.9;
  }

  return value;
}

export function calculateSuggestedPrice(
  purchaseCost: number,
  packagingCost: number,
  rule: PricingRule,
) {
  const totalCost = purchaseCost + packagingCost + rule.extraCost;
  const denominator = 1 - rule.targetMargin - rule.channelFeeRate;
  const rawPrice = denominator > 0 ? totalCost / denominator : totalCost;

  return Number(roundPrice(rawPrice, rule.roundingMode).toFixed(2));
}

export function calculateMinPrice(
  purchaseCost: number,
  packagingCost: number,
  extraCost: number,
  minProfit = 2,
) {
  return Number((purchaseCost + packagingCost + extraCost + minProfit).toFixed(2));
}

export function formatCurrency(value: number) {
  return `¥${value.toFixed(2)}`;
}