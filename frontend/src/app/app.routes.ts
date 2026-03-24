import { Routes } from '@angular/router';
import {GoldMgmt} from './gold-mgmt/gold-mgmt';
import {HomePage} from './home-page/home-page';
import {FundsMgmt} from './funds-mgmt/funds-mgmt';
import {StockMgmt} from './stock-mgmt/stock-mgmt';
import {StockDetail} from './stock-mgmt/stock-detail';

export const routes: Routes = [
  {path: '', component: HomePage},
  {path: 'gold', component: GoldMgmt},
  {path: 'funds', component: FundsMgmt},
  {path: 'stocks', component: StockMgmt},
  {path: 'stocks/:code', component: StockDetail},
];
