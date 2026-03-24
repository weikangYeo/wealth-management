import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {ActivatedRoute, RouterLink} from '@angular/router';

type TxnType = 'BUY' | 'SELL';

interface StockSummary {
  stockCode: string;
  displayName: string;
  dy: number;
  unit: number;
  averagePrice: number;
  realizedGainLoss: number;
  unrealizedGainLoss: number;
  annualizedReturn: number;
}

interface StockTxn {
  id: string;
  txnDate: string;
  txnType: TxnType;
  unit: number;
  unitPrice: number;
  brokerFee: number;
  totalPrice: number;
  remark: string;
}

@Component({
  selector: 'app-stock-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './stock-detail-page.html',
})
export class StockDetail {
  private readonly route = inject(ActivatedRoute);

  protected stockSummary = signal<StockSummary>({
    stockCode: '',
    displayName: '',
    dy: 0,
    unit: 0,
    averagePrice: 0,
    realizedGainLoss: 0,
    unrealizedGainLoss: 0,
    annualizedReturn: 0,
  });

  protected txnDraft = signal({
    txnDate: '',
    txnType: 'BUY' as TxnType,
    unit: 0,
    unitPrice: 0,
    brokerFee: 0,
    remark: '',
  });

  protected transactions = signal<StockTxn[]>([]);

  protected canAddTxn = computed(() => {
    const draft = this.txnDraft();
    return !!draft.txnDate && draft.unit > 0 && draft.unitPrice > 0;
  });

  protected estimatedTotal = computed(() => {
    const draft = this.txnDraft();
    return draft.unit * draft.unitPrice + draft.brokerFee;
  });

  protected totalTxnValue = computed(() =>
    this.transactions().reduce((sum, txn) => sum + txn.totalPrice, 0)
  );

  constructor() {
    this.route.paramMap.subscribe((params) => {
      const stockCode = (params.get('code') ?? 'UNKNOWN').toUpperCase();
      this.stockSummary.set(this.buildMockSummary(stockCode));
      this.transactions.set(this.buildMockTransactions(stockCode));
    });
  }

  protected addTxn() {
    if (!this.canAddTxn()) {
      return;
    }

    const draft = this.txnDraft();
    const newTxn: StockTxn = {
      id: crypto.randomUUID(),
      txnDate: draft.txnDate,
      txnType: draft.txnType,
      unit: Number(draft.unit),
      unitPrice: Number(draft.unitPrice),
      brokerFee: Number(draft.brokerFee),
      totalPrice: this.estimatedTotal(),
      remark: draft.remark.trim(),
    };

    this.transactions.update((txns) => [newTxn, ...txns]);
    this.txnDraft.update((value) => ({
      ...value,
      txnDate: '',
      unit: 0,
      unitPrice: 0,
      brokerFee: 0,
      remark: '',
    }));
  }

  protected updateTxnDate(value: string) {
    this.txnDraft.update((draft) => ({...draft, txnDate: value}));
  }

  protected updateTxnType(value: TxnType) {
    this.txnDraft.update((draft) => ({...draft, txnType: value}));
  }

  protected updateTxnUnit(value: number) {
    this.txnDraft.update((draft) => ({...draft, unit: Number(value)}));
  }

  protected updateTxnUnitPrice(value: number) {
    this.txnDraft.update((draft) => ({...draft, unitPrice: Number(value)}));
  }

  protected updateTxnBrokerFee(value: number) {
    this.txnDraft.update((draft) => ({...draft, brokerFee: Number(value)}));
  }

  protected updateTxnRemark(value: string) {
    this.txnDraft.update((draft) => ({...draft, remark: value}));
  }

  private buildMockSummary(stockCode: string): StockSummary {
    const preset: Record<string, StockSummary> = {
      MAYBANK: {
        stockCode: 'MAYBANK',
        displayName: 'Malayan Banking Berhad',
        dy: 5.84,
        unit: 1200,
        averagePrice: 9.42,
        realizedGainLoss: 1250,
        unrealizedGainLoss: 980,
        annualizedReturn: 7.1,
      },
      TENAGA: {
        stockCode: 'TENAGA',
        displayName: 'Tenaga Nasional Berhad',
        dy: 3.17,
        unit: 600,
        averagePrice: 10.19,
        realizedGainLoss: -150,
        unrealizedGainLoss: 420,
        annualizedReturn: 5.4,
      },
      CIMB: {
        stockCode: 'CIMB',
        displayName: 'CIMB Group Holdings',
        dy: 4.56,
        unit: 950,
        averagePrice: 6.84,
        realizedGainLoss: 300,
        unrealizedGainLoss: -90,
        annualizedReturn: 6.0,
      },
    };

    return (
      preset[stockCode] ?? {
        stockCode,
        displayName: `${stockCode} Holdings`,
        dy: 0,
        unit: 0,
        averagePrice: 0,
        realizedGainLoss: 0,
        unrealizedGainLoss: 0,
        annualizedReturn: 0,
      }
    );
  }

  private buildMockTransactions(stockCode: string): StockTxn[] {
    return [
      {
        id: `${stockCode}-1`,
        txnDate: '2025-11-04',
        txnType: 'BUY',
        unit: 300,
        unitPrice: 9.38,
        brokerFee: 8,
        totalPrice: 2822,
        remark: 'Monthly DCA',
      },
      {
        id: `${stockCode}-2`,
        txnDate: '2025-09-02',
        txnType: 'BUY',
        unit: 200,
        unitPrice: 9.55,
        brokerFee: 8,
        totalPrice: 1918,
        remark: 'Dip buy',
      },
    ];
  }
}

