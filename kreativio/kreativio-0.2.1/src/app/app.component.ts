import { Component, HostBinding, ViewChild, AfterContentInit, OnDestroy, ChangeDetectorRef, ChangeDetectionStrategy } from '@angular/core';
import { OverlayContainer } from '@angular/cdk/overlay';
import { MatDrawer, MatSidenav, MatDialog, MatDialogConfig, DialogPosition } from '@angular/material';
import { takeUntil, take } from 'rxjs/operators'
import { Subject } from 'rxjs';
import { Platform } from '@angular/cdk/platform';
import { ToolbarComponent } from './toolbar/toolbar.component';
import { CartDialog } from './main/cart.component';
import { sidebar_static_link, StaticlinksComponent } from './staticlinks/staticlinks.component';
import { LanguageService } from './services/language.service';
import { PlatformService } from './services/platform.service';
import { Router } from '@angular/router';
import { ResourceService } from './resource.service';
import { CartService } from './cart.service';
@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  //changeDetection: ChangeDetectionStrategy.OnPush
})
export class AppComponent implements AfterContentInit,OnDestroy {
  constructor(cart:CartService){
    let c = cart.parse_cart();
    if(!c){
      cart.init_cart();
    }
  }
  ngAfterContentInit(){}
  ngOnDestroy(){}
}
