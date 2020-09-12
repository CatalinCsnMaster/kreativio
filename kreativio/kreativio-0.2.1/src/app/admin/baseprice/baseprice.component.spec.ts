import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { BasepriceComponent } from './baseprice.component';

describe('BasepriceComponent', () => {
  let component: BasepriceComponent;
  let fixture: ComponentFixture<BasepriceComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ BasepriceComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(BasepriceComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
