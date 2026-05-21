import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';


import {PaginatedResponse} from '../common/pagination.model';
import {Fund, FundTxn, FundTxnReq} from './fund.model';

@Injectable({providedIn: 'root'})
export class FundService {
  private http = inject(HttpClient);
  private readonly FUND_DOMAIN_API_URL = '/funds';

  getAllFundInfo() {
    return this.http.get<PaginatedResponse<Fund>>(this.FUND_DOMAIN_API_URL);
  }

  getFundOverviewByFundCode(fundCode: string) {
    return this.http.get<Fund>(this.FUND_DOMAIN_API_URL + '/' + fundCode);
  }

  getFundTxnByFundCode(fundCode: string) {
    return this.http.get<PaginatedResponse<FundTxn>>(this.FUND_DOMAIN_API_URL + '/' + fundCode + '/transactions');
  }

  addFundInfo(fund: Fund) {
    return this.http.post<{ status: string }>(this.FUND_DOMAIN_API_URL, fund);
  }

  addFundTxn(fundCode: string, fundTxn: FundTxnReq) {
    return this.http.post<{ status: string }>(this.FUND_DOMAIN_API_URL + '/' + fundCode + '/transactions', fundTxn);
  }
}
