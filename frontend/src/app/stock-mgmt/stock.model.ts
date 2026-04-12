export type TxnType = 'BUY' | 'SELL';


export interface CreateStockModel {
  stockName: string;
  displayName: string;
  bursaStockId: number;
}

export interface StockOverview extends CreateStockModel {
  dividendYield: number;
  averagePrice: number;
  unit: number;
  profitLostPercentage: number;
  realizedGainLoss: 0,
  unrealizedGainLoss: 0,
  annualizedReturn: 0,
  // todo fill in another details
}

export interface StockTxn {
  id?: string;
  txnDate: string;
  txnType: TxnType;
  unit: number;
  unitPrice: number;
  brokerFee: number;
  totalPrice: number;
  remark: string;
}

export interface PaginatedResponse<T> {
  content: T[]
}
