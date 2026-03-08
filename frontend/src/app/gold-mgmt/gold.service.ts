import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {GoldTxn} from './gold.model';
import {map} from 'rxjs';

@Injectable({providedIn: 'root'})
export class GoldService {
  private http = inject(HttpClient);
  private readonly GOLD_RESOURCE_API_URL = `/golds`;

  getAllTransactions() {
    return this.http.get<{ golds: GoldTxn[] }>(this.GOLD_RESOURCE_API_URL)
      .pipe(map(res => res.golds))
  }

}
