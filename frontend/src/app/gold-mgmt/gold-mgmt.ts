import {Component, computed, inject, signal} from '@angular/core';
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

  // Computed signal are read-only signals that derive their value from other signals.
  // it is lazy loaded and only calculated when upstream/depended signal done
  protected totalGrams = computed(() =>
    this.goldTransactions().reduce((total, txn) =>
      txn.txnType === 'BUY' ? total + Number(txn.gram) : total - Number(txn.gram), 0
    )
  );

  protected totalPurchased = computed(() =>
    this.goldTransactions().reduce((total, txn) =>
      txn.txnType === 'BUY' ? total + Number(txn.totalPrice) : total - Number(txn.totalPrice), 0
    )
  );

  protected avgPurchasePrice = computed(() => {
    const gram = this.totalGrams();
    return gram > 0 ? this.totalPurchased() / gram : 0;
  });

  ngOnInit() {
    this.goldService.getAllTransactions().subscribe(data => {
      this.goldTransactions.set(data);
    });
  }

}
