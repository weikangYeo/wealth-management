import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {
  CreateDividendModel,
  CreateStockModel,
  Dividend,
  PaginatedResponse,
  StockOverview,
  StockTxn
} from './stock.model';

@Injectable({providedIn: 'root'})
export class StockService {
  private http = inject(HttpClient);
  private readonly STOCK_DOMAIN_API_URL = '/stocks';

  createStock(stock: CreateStockModel) {
    return this.http.post<{ message: string }>(this.STOCK_DOMAIN_API_URL, stock);
  }

  getStocks() {
    return this.http.get<PaginatedResponse<StockOverview>>(this.STOCK_DOMAIN_API_URL);
  }

  getStockOverviewByStockName(stockName: string) {
    return this.http.get<StockOverview>(`${this.STOCK_DOMAIN_API_URL}/${stockName}/overviews`);
  }

  getStockTransactionsByStockName(stockName: string) {
    return this.http.get<PaginatedResponse<StockTxn>>(this.STOCK_DOMAIN_API_URL + `/${stockName}/transactions`);
  }

  createTransaction(stockName: string, txn: StockTxn) {
    return this.http.post(this.STOCK_DOMAIN_API_URL + `/${stockName}/transactions`, txn);
  }

  createDividend(stockName: string, dividend: CreateDividendModel) {
    return this.http.post(this.STOCK_DOMAIN_API_URL + `/${stockName}/dividends`, dividend);
  }

  getDividends(stockName: string) {
    return this.http.get<PaginatedResponse<Dividend>>(this.STOCK_DOMAIN_API_URL + `/${stockName}/dividends`);
  }
}
