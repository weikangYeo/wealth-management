export type TxnType = 'BUY' | 'SELL';


export interface CreateStockModel {
  stockName: string;
  displayName: string;
  bursaStockId: string;
}

export interface StockOverview extends CreateStockModel {
  averagePrice: number;
  totalUnit: number;
  realizedGainLoss: number | null;
  unrealizedGainLoss: number;
  annualizedReturn: number | null;
  dividendPayback: number;
  dividendYield: number | null;
  totalDividend: number | null;
  roi: number | null;
}

export interface StockTxn {
  id?: string;
  txnDate: string;
  txnType: TxnType;
  unit: number;
  unitPrice: number;
  brokerFee: number;
  totalAmount?: number;
  remark: string;
}

export interface CreateDividendModel {
  exDate: string;
  paymentDate: string;
  stockUnit: number;
  dividendPerUnit: number;
  taxPercentage: number;
  remark: string;
}

export interface Dividend {
  stockName: string;
  exDate: string;
  paymentDate: string;
  stockUnit: number;
  dividendPerUnit: number;
  taxPercentage: number;
  grossAmount: number;
  netAmount: number;
  remark: string;
}
