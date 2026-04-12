import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {CreateStockModel, PaginatedResponse, StockOverview, StockTxn} from './stock.model';

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

  getStockOverviewByStockCode(stockCode: string) {
    return this.http.get<StockOverview>(`${this.STOCK_DOMAIN_API_URL}/${stockCode}/overviews`);
  }

  getStockTransactions(stockCode: string) {
    return this.http.get<PaginatedResponse<StockTxn>>(this.STOCK_DOMAIN_API_URL + `/${stockCode}/transactions`);
  }

  createTransaction(stockCode: string, txn: StockTxn) {
    return this.http.post(this.STOCK_DOMAIN_API_URL + `/${stockCode}/transactions`, txn);
  }
}
