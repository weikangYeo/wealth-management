import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {Router} from '@angular/router';
import {CreateStockModel, StockAggregatedInfo} from './stock.model';
import {StockService} from './stock.service';
import {MatSnackBar} from '@angular/material/snack-bar';

@Component({
  selector: 'app-stock-mgmt',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './stock-mgmt-page.html',
})
export class StockMgmt {
  private readonly router = inject(Router);
  private stockService = inject(StockService);
  private snackBar = inject(MatSnackBar);

  protected createStockReq = signal<CreateStockModel>({
    stockCode: '',
    displayName: '',
  });

  protected activeStock = signal<CreateStockModel | null>(null);

  protected stocks = signal<StockAggregatedInfo[]>([]);

  protected canSaveStock = computed(() => {
    const draft = this.createStockReq();
    return draft.stockCode.trim().length > 0 && draft.displayName.trim().length > 0;
  });

  protected stockCount = computed(() =>
    this.stocks().length
  );

  protected saveStockProfile() {
    if (!this.canSaveStock()) {
      return;
    }

    const draft = this.createStockReq();
    const stockCode = draft.stockCode.trim().toUpperCase();
    const displayName = draft.displayName.trim();
    this.activeStock.set({stockCode, displayName});
    this.stockService.createStock(draft).subscribe({
      next: () => {
        this.snackBar.open('Stock Code Added', 'OK', {
          duration: 3000,
          verticalPosition: 'top'
        });
        this.fetchAllStockProfile();
      }, error: (error) => {
        console.error('Import failed:', error);
      }
    });
  }

  protected openStockDetails(stockCode: string) {
    this.router.navigate(['/stocks', stockCode]);
  }

  protected updateStockCodeRequest(value: string) {
    this.createStockReq.update((draft) => ({...draft, stockCode: value}));
  }

  protected updateStockDisplayNameReq(value: string) {
    this.createStockReq.update((draft) => ({...draft, displayName: value}));
  }

  private fetchAllStockProfile(){
    this.stockService.getStocks().subscribe(data => {
      this.stocks.set(data.content)
    })
  }

  ngOnInit() {
    this.fetchAllStockProfile();
  }

}
