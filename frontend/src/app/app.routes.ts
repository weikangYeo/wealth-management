import { Routes } from '@angular/router';
import {GoldMgmt} from './gold-mgmt/gold-mgmt';
import {HomePage} from './home-page/home-page';
import {FundsMgmt} from './funds-mgmt/funds-mgmt';

export const routes: Routes = [
  {path: '', component: HomePage},
  {path: 'gold', component: GoldMgmt},
  {path: 'funds', component: FundsMgmt},
];

