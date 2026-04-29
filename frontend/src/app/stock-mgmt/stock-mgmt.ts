import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {Router} from '@angular/router';
import {CreateStockModel, StockOverview} from './stock.model';
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

  protected activeStock = signal<CreateStockModel | null>(null);

  protected stocks = signal<StockOverview[]>([]);

  protected stockCount = computed(() =>
    this.stocks().length
  );

  protected saveStockProfile(stockName: string, displayName: string, bursaStockId: string) {

    if (!stockName || !displayName || !bursaStockId) {
      this.snackBar.open('Missing mandatory field to create stock code', 'OK', {
        duration: 3000,
        verticalPosition: 'top'
      });
      return;
    }
    const draft: CreateStockModel = {
      stockName: stockName,
      displayName: displayName,
      bursaStockId: bursaStockId,
    }
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

  protected openStockDetails(stockName: string) {
    this.router.navigate(['/stocks', stockName]);
  }

  private fetchAllStockProfile() {
    this.stockService.getStocks().subscribe(data => {
      this.stocks.set(data.content);
    });
  }

  ngOnInit() {
    this.fetchAllStockProfile();
  }

}
