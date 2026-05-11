export type FundTxnType = 'BUY' | 'SELL';

export interface Fund {
  fundCode: string;
  name: string;
  url: string;
}

export interface FundTxn {
  id?: string;
  fundCode: string;
  txnDate: string;
  unit: number;
  unitPrice: number;
  salesCharge: number;
  grossTotalPrice?: number;
  netTotalPrice?: number;
  txnType: FundTxnType;
  remark: string;
}

export interface FundPriceHistory {
  fundCode: string;
  priceDate: string;
  lastDonePrice: number;
}

export interface PaginatedResponse<T> {
  content: T[];
}
