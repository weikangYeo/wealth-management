import {ChangeDetectionStrategy, Component, computed, inject, signal} from '@angular/core';
import {GoldTxn} from './gold.model';
import {GoldService} from './gold.service';
import {MatButtonModule} from '@angular/material/button';
import {MatIconModule} from '@angular/material/icon';
import {MatFormFieldModule} from '@angular/material/form-field';
import {MatInputModule} from '@angular/material/input';
import {MatSnackBar, MatSnackBarModule} from '@angular/material/snack-bar';
import {CommonModule} from '@angular/common';
import {MatTooltipModule} from '@angular/material/tooltip';

@Component({
  selector: 'app-gold-mgmt',
  standalone: true,
  imports: [
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatButtonModule,
    CommonModule,
    MatSnackBarModule,
    MatTooltipModule
  ],
  changeDetection: ChangeDetectionStrategy.Default, // or OnPush
  templateUrl: "./gold-mgmt-page.html",
  styles: [],
})
export class GoldMgmt {

  private snackBar = inject(MatSnackBar);
  private goldService = inject(GoldService);
  protected goldTransactions = signal<GoldTxn[]>([]);
  protected latestUnitPrice = signal<number>(0);
  protected latestPriceDate = signal<string>("");
  selectedFile: File | null = null;
  fileName: string = '';

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

  protected latestGoldsValue = computed(() => {
    const totalGram = this.totalGrams();
    const latestUnitPrice = this.latestUnitPrice();
    return Number(latestUnitPrice) * Number(totalGram);
  });

  protected unrealizedGainLoss = computed(() => {
    const totalCost = this.totalPurchased();
    const latestValue = this.latestGoldsValue();
    return Number(latestValue) / Number(totalCost);
  });

  ngOnInit() {
    this.fetchAllTransactions();
    this.fetchLatestPrice();
  }

  private fetchAllTransactions() {
    this.goldService.getAllTransactions().subscribe(data => {
      this.goldTransactions.set(data);
    });
  }

  private fetchLatestPrice() {
    this.goldService.getLatestPrice().subscribe({
      next: value => {
        this.latestUnitPrice.set(value.latestPrice);
        this.latestPriceDate.set(value.date);
      },
      error: err => console.log('Error fetching latest price: ', err)
    });
  }

  onFileSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.selectedFile = input.files[0];
      this.fileName = this.selectedFile.name;
    }
  }


  onBulkImport() {
    if (!this.selectedFile) {
      alert('Please select a file');
      return;
    }
    this.goldService.bulkImportGolds(this.selectedFile).subscribe({
      next: (response) => {
        this.goldService.getAllTransactions().subscribe(data => {
          this.goldTransactions.set(data);
          this.snackBar.open('Bulk Import Successful', 'OK', {
            duration: 3000,
            verticalPosition: 'top'
          });
        });
      },
      error: (error) => {
        console.error('Import failed:', error);
      }
    });
  }
}
