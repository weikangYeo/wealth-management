export type FundTxnType = 'BUY' | 'SELL';

export interface Fund {
  fundCode: string;
  name: string;
  url: string;
}

export interface FundTxn extends FundTxnReq{
  id?: string;
  fundCode: string;
  netInvestmentAmount?: number;
  totalAmount?: number;
}

export interface FundTxnReq {
  txnDate: string;
  unit: number;
  unitPrice: number;
  salesCharge: number;
  txnType: FundTxnType;
  remark: string;
}

export interface FundPriceHistory {
  fundCode: string;
  priceDate: string;
  nav: number;
}

export interface PaginatedResponse<T> {
  content: T[];
}
