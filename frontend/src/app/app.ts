import {Component, signal} from '@angular/core';
import {RouterLink, RouterOutlet} from '@angular/router';
import {GoldMgmt} from './gold-mgmt/gold-mgmt';


@Component({
  selector: 'app-root',
  imports: [RouterOutlet, RouterLink, GoldMgmt],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  protected readonly title = signal('frontend');
}
