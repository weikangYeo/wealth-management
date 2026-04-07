import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {Router} from '@angular/router';
import {CreateStockModel} from './stock.model';
import {StockService} from './stock.service';
import {MatSnackBar} from '@angular/material/snack-bar';

interface StockCard {
  stockCode: string;
  displayName: string;
  dy: number;
  avgPrice: number;
  unit: number;
}

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

  protected stocks = signal<StockCard[]>([
    {stockCode: 'MAYBANK', displayName: 'Malayan Banking Berhad', dy: 5.84, avgPrice: 9.42, unit: 1200},
    {stockCode: 'TENAGA', displayName: 'Tenaga Nasional Berhad', dy: 3.17, avgPrice: 10.19, unit: 600},
    {stockCode: 'CIMB', displayName: 'CIMB Group Holdings', dy: 4.56, avgPrice: 6.84, unit: 950},
  ]);

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
        // todo recall GET api and reload stock list
      }, error: (error) => {
        console.error('Import failed:', error);
      }
    });

    // this.stocks.update((items) => {
    //   const index = items.findIndex((item) => item.stockCode === stockCode);
    //   if (index >= 0) {
    //     const next = [...items];
    //     next[index] = {...next[index], displayName};
    //     return next;
    //   }
    //
    //   return [
    //     {
    //       stockCode,
    //       displayName,
    //       dy: 0,
    //       avgPrice: 0,
    //       unit: 0,
    //     },
    //     ...items,
    //   ];
    // });
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

}
