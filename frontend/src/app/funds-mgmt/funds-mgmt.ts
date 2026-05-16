import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { Fund } from './fund.model';
import {FundService} from './fund.service';

@Component({
  selector: 'app-funds-mgmt',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './funds-mgmt-page.html',
})
export class FundsMgmt {
  private readonly router = inject(Router);
  private fundService = inject(FundService);

  protected funds = signal<Fund[]>([]);
  protected fundCount = computed(() => this.funds().length);

  protected saveFundProfile(fundCode: string, name: string, url: string) {
    if (!fundCode || !name) {
      return;
    }
    // todo: wire up service
    this.funds.update(list => [...list, { fundCode, name, url }]);
  }

  protected openFundDetails(fundCode: string) {
    this.router.navigate(['/funds', fundCode]);
  }

  private getAllFundInfo(){
    this.fundService.getAllFundInfo().subscribe(data => {
      this.funds.set(data.content)
    })
  }

  ngOnInit() {
    this.getAllFundInfo();

  }
}
