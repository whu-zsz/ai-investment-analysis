export interface NumericAxisRangeOptions {
  minPaddingRatio?: number;
  maxPaddingRatio?: number;
  minPaddingAbs?: number;
  includeZero?: boolean;
  symmetricAroundZero?: boolean;
  roundMode?: 'none' | 'step' | 'magnitude';
  splitNumber?: number;
}

function ensureFiniteValues(values: number[]) {
  return values.filter((value) => Number.isFinite(value));
}

function computeNiceStep(span: number, splitNumber = 4) {
  const safeSpan = Math.max(Math.abs(span), Number.EPSILON);
  const roughStep = safeSpan / Math.max(splitNumber, 1);
  const magnitude = 10 ** Math.floor(Math.log10(roughStep));
  const residual = roughStep / magnitude;

  if (residual <= 1) {
    return magnitude;
  }
  if (residual <= 2) {
    return 2 * magnitude;
  }
  if (residual <= 5) {
    return 5 * magnitude;
  }
  return 10 * magnitude;
}

function computeMagnitudeStep(value: number) {
  const safeValue = Math.max(Math.abs(value), Number.EPSILON);
  return 10 ** (Math.floor(Math.log10(safeValue)) - 1);
}

function normalizeAxisRange(min: number, max: number, roundMode: NumericAxisRangeOptions['roundMode'] = 'none', splitNumber = 4) {
  if (roundMode === 'none') {
    return { min, max };
  }

  const spanStep = computeNiceStep(max - min, splitNumber);
  const magnitudeStep = computeMagnitudeStep(Math.max(Math.abs(min), Math.abs(max)));
  const interval = roundMode === 'magnitude' ? Math.max(spanStep, magnitudeStep) : spanStep;

  return {
    min: Math.floor(min / interval) * interval,
    max: Math.ceil(max / interval) * interval,
    interval,
  };
}

export function computeNumericAxisRange(values: number[], options: NumericAxisRangeOptions = {}) {
  const finiteValues = ensureFiniteValues(values);
  if (!finiteValues.length) {
    return undefined;
  }

  const {
    minPaddingRatio = 0.06,
    maxPaddingRatio = minPaddingRatio,
    minPaddingAbs = 1,
    includeZero = false,
    symmetricAroundZero = false,
    roundMode = 'none',
    splitNumber = 4,
  } = options;

  let min = Math.min(...finiteValues);
  let max = Math.max(...finiteValues);

  if (includeZero) {
    min = Math.min(min, 0);
    max = Math.max(max, 0);
  }

  if (symmetricAroundZero) {
    const absMax = Math.max(Math.abs(min), Math.abs(max), minPaddingAbs);
    const padding = Math.max(absMax * maxPaddingRatio, minPaddingAbs);
    return normalizeAxisRange(-(absMax + padding), absMax + padding, roundMode, splitNumber);
  }

  const span = max - min;
  const baseSpan = span === 0 ? Math.max(Math.abs(max), Math.abs(min), minPaddingAbs) : span;
  const minPadding = Math.max(baseSpan * minPaddingRatio, minPaddingAbs);
  const maxPadding = Math.max(baseSpan * maxPaddingRatio, minPaddingAbs);

  return normalizeAxisRange(min - minPadding, max + maxPadding, roundMode, splitNumber);
}

export function computeZeroAwareAxisRange(values: number[], options: Omit<NumericAxisRangeOptions, 'includeZero' | 'symmetricAroundZero'> = {}) {
  const finiteValues = ensureFiniteValues(values);
  if (!finiteValues.length) {
    return undefined;
  }

  const hasPositive = finiteValues.some((value) => value > 0);
  const hasNegative = finiteValues.some((value) => value < 0);

  return computeNumericAxisRange(finiteValues, {
    ...options,
    includeZero: true,
    symmetricAroundZero: hasPositive && hasNegative,
  });
}

export function computeCandleAxisRange(
  highValues: number[],
  lowValues: number[],
  options: Pick<NumericAxisRangeOptions, 'roundMode' | 'splitNumber'> & { paddingRatio?: number; minPaddingAbs?: number } = {},
) {
  const highs = ensureFiniteValues(highValues);
  const lows = ensureFiniteValues(lowValues);
  if (!highs.length || !lows.length) {
    return undefined;
  }

  const {
    paddingRatio = 0.02,
    minPaddingAbs = 1,
    roundMode = 'none',
    splitNumber = 4,
  } = options;

  const max = Math.max(...highs);
  const min = Math.min(...lows);
  const span = max - min;
  const baseSpan = span === 0 ? Math.max(Math.abs(max), Math.abs(min), minPaddingAbs) : span;
  const padding = Math.max(baseSpan * paddingRatio, minPaddingAbs);

  return normalizeAxisRange(min - padding, max + padding, roundMode, splitNumber);
}
