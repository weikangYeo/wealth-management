import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {ActivatedRoute, RouterLink} from '@angular/router';
import {CreateDividendModel, Dividend, StockOverview, StockTxn, TxnType} from './stock.model';
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

  protected stockName = signal<String>('');
  protected stockOverview = signal<StockOverview>({
    stockName: '',
    displayName: '',
    bursaStockId: 0,
    unit: 0,
    averagePrice: 0,
    realizedGainLoss: 0,
    unrealizedGainLoss: 0,
    annualizedReturn: 0,
    dividendYield: 0,
    profitLostPercentage: 0
  });

  // signal to control the add stock/dividend tab
  protected addTab = signal<'transaction' | 'dividend'>('transaction');
  // signal to control list of stock/dividend tab
  protected listTab = signal<'transaction' | 'dividend'>('transaction');
  protected transactions = signal<StockTxn[]>([]);
  protected dividends = signal<Dividend[]>([]);
  protected totalTxnValue = computed(() =>
    this.transactions().reduce(
      (sum, txn) => sum + (txn.totalPrice || 0), 0)
  );
  // todo shall i move to another component?
  // dividend tab inputs
  protected exDate = signal<string>('');
  protected paymentDate = signal<string>('');
  protected stockUnit = signal<number>(0);
  protected dividendPerUnit = signal<number>(0);
  protected grossAmount = computed(() => this.dividendPerUnit() * this.stockUnit());
  protected withHoldingTax = signal<number>(0);
  protected netDividendAmount = computed(() => this.grossAmount() - this.withHoldingTax());
  protected remark = signal<string>('');

  ngOnInit() {
    this.route.paramMap.subscribe((params) => {
      const stockName = (params.get('code') ?? 'UNKNOWN').toUpperCase();
      this.stockName.set(stockName);
      this.loadStockOverview(stockName);
      this.loadStockTransactions(stockName);
      this.loadDividends(stockName);
    });
  }

  // method 1, pass value via html, so no signal variable required
  protected addTxn(
    txnDate: string,
    txnType: string,
    unit: string,
    unitPrice: string,
    brokerFee: string,
    remark: string
  ) {

    // Validate required fields
    if (!txnDate || !unit) {
      this.snackBar.open('Date or Unit is empty', 'OK', {
        duration: 3000,
        verticalPosition: 'top'
      });
      return;
    }

    const unitNum = Number(unit);
    const unitPriceNum = Number(unitPrice) || 0;
    const brokerFeeNum = Number(brokerFee) || 0;

    if (unitNum <= 0) {
      return;
    }

    const newTxn: StockTxn = {
      txnDate,
      txnType: txnType as TxnType,
      unit: unitNum,
      unitPrice: unitPriceNum,
      brokerFee: brokerFeeNum,
      remark: remark?.trim() || '',
    };

    this.stockService.createTransaction(this.stockName().toString(), newTxn).subscribe({
      next: () => {
        this.snackBar.open('Create Transaction Successful', 'OK', {
          duration: 3000,
          verticalPosition: 'top'
        });
        this.loadStockTransactions(this.stockName().toString());
      }, error: (error) => {
        console.log(error);
      }
    });
  }

  // method 2, hold signal variable in ts.
  protected addDividend() {
    let dividend: CreateDividendModel;
    dividend = {
      exDate: this.exDate(),
      paymentDate: this.paymentDate(),
      stockUnit: this.stockUnit(),
      dividendPerUnit: this.dividendPerUnit(),
      taxPercentage: this.withHoldingTax(),
      remark: this.remark()
    };

    this.stockService.createDividend(this.stockName().toString(), dividend).subscribe({
      next: () => {
        this.snackBar.open('Create Dividend Successful', 'OK', {
          duration: 3000,
          verticalPosition: 'top'
        });
        this.loadDividends(this.stockName().toString());
      },
      error: (error) => {
        console.log(error);
      }
    });
  }

  private loadStockOverview(stockName: string) {
    this.stockService.getStockOverviewByStockName(stockName).subscribe(data => {
      this.stockOverview.set(data);
    });
  }

  private loadStockTransactions(stockName: string) {
    this.stockService.getStockTransactionsByStockName(stockName).subscribe(data => {
      this.transactions.set(data.content);
    });
  }

  private loadDividends(stockName: string) {
    this.stockService.getDividends(stockName).subscribe(data => {
      this.dividends.set(data.content);
    })
  }

  protected readonly Number = Number;
}

