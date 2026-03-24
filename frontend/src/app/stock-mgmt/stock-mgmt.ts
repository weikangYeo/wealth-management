import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {Router} from '@angular/router';

interface StockProfile {
  stockCode: string;
  displayName: string;
}

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

  protected profileDraft = signal<StockProfile>({
    stockCode: '',
    displayName: '',
  });

  protected activeStock = signal<StockProfile | null>(null);

  protected stocks = signal<StockCard[]>([
    {stockCode: 'MAYBANK', displayName: 'Malayan Banking Berhad', dy: 5.84, avgPrice: 9.42, unit: 1200},
    {stockCode: 'TENAGA', displayName: 'Tenaga Nasional Berhad', dy: 3.17, avgPrice: 10.19, unit: 600},
    {stockCode: 'CIMB', displayName: 'CIMB Group Holdings', dy: 4.56, avgPrice: 6.84, unit: 950},
  ]);

  protected canSaveStock = computed(() => {
    const draft = this.profileDraft();
    return draft.stockCode.trim().length > 0 && draft.displayName.trim().length > 0;
  });

  protected stockCount = computed(() =>
    this.stocks().length
  );

  protected saveStockProfile() {
    if (!this.canSaveStock()) {
      return;
    }

    const draft = this.profileDraft();
    const stockCode = draft.stockCode.trim().toUpperCase();
    const displayName = draft.displayName.trim();
    this.activeStock.set({stockCode, displayName});

    this.stocks.update((items) => {
      const index = items.findIndex((item) => item.stockCode === stockCode);
      if (index >= 0) {
        const next = [...items];
        next[index] = {...next[index], displayName};
        return next;
      }

      return [
        {
          stockCode,
          displayName,
          dy: 0,
          avgPrice: 0,
          unit: 0,
        },
        ...items,
      ];
    });
  }

  protected openStockDetails(stockCode: string) {
    this.router.navigate(['/stocks', stockCode]);
  }

  protected updateProfileStockCode(value: string) {
    this.profileDraft.update((draft) => ({...draft, stockCode: value}));
  }

  protected updateProfileDisplayName(value: string) {
    this.profileDraft.update((draft) => ({...draft, displayName: value}));
  }

}
