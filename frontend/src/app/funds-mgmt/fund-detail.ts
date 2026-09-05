import {CommonModule} from '@angular/common';
import {Component, computed, inject, signal} from '@angular/core';
import {ActivatedRoute, RouterLink} from '@angular/router';
import {Fund, FundTxn, FundTxnType} from './fund.model';
import {FundService} from './fund.service';

@Component({
  selector: 'app-fund-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './fund-detail-page.html',
})
export class FundDetail {
  private readonly route = inject(ActivatedRoute);
  private readonly fundService = inject(FundService);

  protected fund = signal<Fund>({fundCode: '', name: '', fsmUrl: ''});
  protected transactions = signal<FundTxn[]>([]);
  protected fundCode = signal<string>('');

  protected txnCount = computed(() => this.transactions().length);

  protected totalUnits = computed(() =>
    this.transactions().reduce((sum, t) => {
      return t.txnType === 'BUY' ? sum + t.unit : sum - t.unit;
    }, 0)
  );

  protected totalNetInvestmentAmount = computed(() =>
    this.transactions().reduce((sum, t) => sum + (t.netInvestmentAmount ?? 0), 0)
  );

  protected totalSalesCharge = computed(() =>
    this.transactions().reduce((sum, t) => sum + t.salesCharge, 0)
  );

  protected avgUnitPrice = computed(() => {
    const units = this.totalUnits();
    const netInvestmentAmount = this.transactions().reduce((sum, t) => sum + (t.netInvestmentAmount ?? 0), 0);
    return units > 0 ? netInvestmentAmount / units : 0;
  });

  // form preview signals
  protected netInvestmentPreview = signal<number>(0);
  protected totalAmountPreview = signal<number>(0);

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const fundCode = (params.get('code') ?? '').toUpperCase();
      this.fundCode.set(fundCode);
      this.loadFundInfo(fundCode)
      this.loadFundTxns(fundCode)
    });
  }

  private loadFundInfo(fundCode:string){
    this.fundService.getFundOverviewByFundCode(fundCode).subscribe({
      next: data => {
        this.fund.set(data)
      },
      error: err => console.log(err)
    })
  }

  private loadFundTxns(fundCode: string) {
    this.fundService.getFundTxnByFundCode(fundCode).subscribe(
      {
        next: data => {
          this.transactions.set(data.content ?? [])
        },
        error : err => console.log(err)
      }
    );
  }

  protected onUnitInput(unit: string, unitPrice: string) {
    const netInvestmentAmount = (Number(unit) || 0) * (Number(unitPrice) || 0);
    this.netInvestmentPreview.set(netInvestmentAmount);
    this.totalAmountPreview.set(netInvestmentAmount);
  }

  protected onSalesChargeInput(unit: string, unitPrice: string, salesCharge: string) {
    const netInvestmentAmount = (Number(unit) || 0) * (Number(unitPrice) || 0);
    this.netInvestmentPreview.set(netInvestmentAmount);
    this.totalAmountPreview.set(netInvestmentAmount + (Number(salesCharge) || 0));
  }

  protected resetPreview() {
    this.netInvestmentPreview.set(0);
    this.totalAmountPreview.set(0);
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
    const netInvestmentAmount = unitNum * unitPriceNum;
    const totalAmount = netInvestmentAmount + salesChargeNum;

    if (unitNum <= 0) return;

    const newTxn: FundTxn = {
      id: crypto.randomUUID(),
      fundCode: this.fund().fundCode,
      txnDate,
      txnType: txnType as FundTxnType,
      unit: unitNum,
      unitPrice: unitPriceNum,
      salesCharge: salesChargeNum,
      netInvestmentAmount: netInvestmentAmount,
      totalAmount: totalAmount,
      remark: remark?.trim() || '',
    };

    // todo: wire up service
    this.transactions.update(list => [...list, newTxn]);
  }
}
