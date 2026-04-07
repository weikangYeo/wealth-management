import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {CreateStockModel, StockAggregatedInfo} from './stock.model';

@Injectable({providedIn: 'root'})
export class StockService {
  private http = inject(HttpClient);
  private readonly STOCK_DOMAIN_API_URL = '/stocks';

  createStock(stock: CreateStockModel) {
    return this.http.post<{ message: string }>(this.STOCK_DOMAIN_API_URL, stock);
  }

  getStocks() {
    return this.http.get<StockAggregatedInfo>(this.STOCK_DOMAIN_API_URL);
  }
}
