import {Component, inject, signal} from '@angular/core';
import {GoldTxn} from './gold.model';
import {GoldService} from './gold.service';

@Component({
  selector: 'app-gold-mgmt',
  standalone: true,
  imports: [],
  templateUrl: "./gold-mgmt-page.html",
  styles: [],
})
export class GoldMgmt {

  private goldService = inject(GoldService);
  protected goldTransactions = signal<GoldTxn[]>([]);

  ngOnInit(){
    this.goldService.getAllTransactions().subscribe( data => {
      this.goldTransactions.set(data);
    })
  }

}
