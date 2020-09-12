import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { StaticlinksComponent } from './staticlinks.component';

describe('StaticlinksComponent', () => {
  let component: StaticlinksComponent;
  let fixture: ComponentFixture<StaticlinksComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ StaticlinksComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(StaticlinksComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
