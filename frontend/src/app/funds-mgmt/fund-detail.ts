import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { Fund, FundTxn, FundTxnType } from './fund.model';

@Component({
  selector: 'app-fund-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './fund-detail-page.html',
})
export class FundDetail {
  private readonly route = inject(ActivatedRoute);

  protected fund = signal<Fund>({ fundCode: '', name: '', url: '' });
  protected transactions = signal<FundTxn[]>([]);

  protected txnCount = computed(() => this.transactions().length);

  protected totalUnits = computed(() =>
    this.transactions().reduce((sum, t) => {
      return t.txnType === 'BUY' ? sum + t.unit : sum - t.unit;
    }, 0)
  );

  protected totalNetInvested = computed(() =>
    this.transactions().reduce((sum, t) => sum + (t.netTotalPrice ?? 0), 0)
  );

  protected totalSalesCharge = computed(() =>
    this.transactions().reduce((sum, t) => sum + t.salesCharge, 0)
  );

  protected avgUnitPrice = computed(() => {
    const units = this.totalUnits();
    const gross = this.transactions().reduce((sum, t) => sum + (t.grossTotalPrice ?? 0), 0);
    return units > 0 ? gross / units : 0;
  });

  // form preview signals
  protected grossPreview = signal<number>(0);
  protected netPreview = signal<number>(0);

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const fundCode = (params.get('code') ?? '').toUpperCase();
      // todo: wire up service to load fund and transactions
      this.fund.set({ fundCode, name: fundCode, url: '' });
    });
  }

  protected onUnitInput(unit: string, unitPrice: string) {
    const gross = (Number(unit) || 0) * (Number(unitPrice) || 0);
    this.grossPreview.set(gross);
    this.netPreview.set(gross);
  }

  protected onSalesChargeInput(unit: string, unitPrice: string, salesCharge: string) {
    const gross = (Number(unit) || 0) * (Number(unitPrice) || 0);
    this.grossPreview.set(gross);
    this.netPreview.set(gross + (Number(salesCharge) || 0));
  }

  protected resetPreview() {
    this.grossPreview.set(0);
    this.netPreview.set(0);
  }

  protected addTxn(
    txnDate: string,
    txnType: string,
    unit: string,
    unitPrice: string,
    salesCharge: string,
    remark: string
  ) {
    if (!txnDate || !unit || !unitPrice) {
      return;
    }

    const unitNum = Number(unit);
    const unitPriceNum = Number(unitPrice);
    const salesChargeNum = Number(salesCharge) || 0;
    const gross = unitNum * unitPriceNum;
    const net = gross + salesChargeNum;

    if (unitNum <= 0) return;

    const newTxn: FundTxn = {
      id: crypto.randomUUID(),
      fundCode: this.fund().fundCode,
      txnDate,
      txnType: txnType as FundTxnType,
      unit: unitNum,
      unitPrice: unitPriceNum,
      salesCharge: salesChargeNum,
      grossTotalPrice: gross,
      netTotalPrice: net,
      remark: remark?.trim() || '',
    };

    // todo: wire up service
    this.transactions.update(list => [...list, newTxn]);
  }
}
