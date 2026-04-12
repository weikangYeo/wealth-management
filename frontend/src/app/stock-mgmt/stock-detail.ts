import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {ActivatedRoute, RouterLink} from '@angular/router';
import {StockOverview, StockTxn, TxnType} from './stock.model';
import {StockService} from './stock.service';
import {MatSnackBar} from '@angular/material/snack-bar';

@Component({
  selector: 'app-stock-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './stock-detail-page.html',
})
export class StockDetail {
  private readonly route = inject(ActivatedRoute);
  private stockService = inject(StockService);
  private snackBar = inject(MatSnackBar);

  protected stockCode = signal<String>('');
  protected stockOverview = signal<StockOverview>({
    stockCode: '',
    displayName: '',
    unit: 0,
    averagePrice: 0,
    realizedGainLoss: 0,
    unrealizedGainLoss: 0,
    annualizedReturn: 0,
    dividendYield: 0,
    profitLostPercentage: 0
  });

  protected transactions = signal<StockTxn[]>([]);

  protected totalTxnValue = computed(() =>
    this.transactions().reduce(
      (sum, txn) => sum + txn.totalPrice, 0)
  );

  ngOnInit() {
    this.route.paramMap.subscribe((params) => {
      const stockCode = (params.get('code') ?? 'UNKNOWN').toUpperCase();
      this.stockCode.set(stockCode);
      this.loadStockOverview(stockCode);
      this.loadStockTransactions(stockCode);
    });
  }

  protected addTxn(
    txnDate: string,
    txnType: string,
    unit: string,
    unitPrice: string,
    brokerFee: string,
    remark: string
  ) {

    // Validate required fields
    if (!txnDate || !unit ) {
      this.snackBar.open('Date or Unit is empty', 'OK', {
        duration: 3000,
        verticalPosition: 'top'
      });
      return;
    }

    const unitNum = Number(unit);
    const unitPriceNum = Number(unitPrice) || 0;
    const brokerFeeNum = Number(brokerFee) || 0;

    if (unitNum <= 0 ) {
      return;
    }

    const newTxn: StockTxn = {
      id: crypto.randomUUID(),
      txnDate,
      txnType: txnType as TxnType,
      unit: unitNum,
      unitPrice: unitPriceNum,
      brokerFee: brokerFeeNum,
      totalPrice: unitNum * unitPriceNum + brokerFeeNum,
      remark: remark?.trim() || '',
    };

    this.stockService.createTransaction(this.stockCode().toString(), newTxn).subscribe(data => {
      this.snackBar.open('Create Transaction Successful', 'OK', {
        duration: 3000,
        verticalPosition: 'top'
      });
      this.loadStockTransactions(this.stockCode().toString());
    });
  }

  private loadStockOverview(stockCode: string) {
    this.stockService.getStockOverviewByStockCode(stockCode).subscribe(data => {
      this.stockOverview.set(data);
    });
  }

  private loadStockTransactions(stockCode: string) {
    this.stockService.getStockTransactions(stockCode).subscribe(data => {
      this.transactions.set(data.content);
    });
  }

  protected readonly Number = Number;
}

