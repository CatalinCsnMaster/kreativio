import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AListComponent } from './alist.component';

describe('ListComponent', () => {
  let component: AListComponent;
  let fixture: ComponentFixture<AListComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AListComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
