export type Tone = 'positive' | 'warning' | 'neutral';

export interface StatItem {
  label: string;
  value: string;
  hint: string;
  tone: Tone;
}
