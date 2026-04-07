export interface CreateStockModel {
  stockCode: string;
  displayName: string;
}

export interface StockAggregatedInfo extends CreateStockModel {
  dividendYield: number;
  avgPrice: number;
  unit: number;
  profitLostPercentage: number;
  // todo fill in another details
}
