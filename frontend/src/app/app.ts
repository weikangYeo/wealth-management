import {Component, signal} from '@angular/core';
import {RouterLink, RouterOutlet} from '@angular/router';
import {GoldMgmt} from './gold-mgmt/gold-mgmt';
import {HomePage} from './home-page/home-page';
import {FundsMgmt} from './funds-mgmt/funds-mgmt';


@Component({
  selector: 'app-root',
  imports: [RouterOutlet, RouterLink, GoldMgmt, HomePage, FundsMgmt],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  protected readonly title = signal('frontend');
}

