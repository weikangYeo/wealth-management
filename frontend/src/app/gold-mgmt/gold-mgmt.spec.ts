import { ComponentFixture, TestBed } from '@angular/core/testing';

import { GoldMgmt } from './gold-mgmt';

describe('GoldMgmt', () => {
  let component: GoldMgmt;
  let fixture: ComponentFixture<GoldMgmt>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GoldMgmt]
    })
    .compileComponents();

    fixture = TestBed.createComponent(GoldMgmt);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
