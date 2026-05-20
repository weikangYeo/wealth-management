import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';


import {PaginatedResponse} from '../common/pagination.model';
import {Fund} from './fund.model';

@Injectable({providedIn: 'root'})
export class FundService {
  private http = inject(HttpClient);
  private readonly FUND_DOMAIN_API_URL = '/funds';

  getAllFundInfo() {
    return this.http.get<PaginatedResponse<Fund>>(this.FUND_DOMAIN_API_URL);
  }

  addFundInfo(fund:Fund) {
    return this.http.post<{status: string}>(this.FUND_DOMAIN_API_URL, fund)
  }
}
