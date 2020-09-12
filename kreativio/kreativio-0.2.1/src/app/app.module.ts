import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { MorphModule } from 'schema-admin';
import { MaterialModule } from './material/material.module';
import { ToolbarComponent } from './toolbar/toolbar.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { OverlayModule } from '@angular/cdk/overlay';
import { MainComponent } from './main/main.component';
import { AdminModule } from './admin/admin.module';
import { MDBBootstrapModule } from 'angular-bootstrap-md';
import { HttpClientModule } from '@angular/common/http';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';
import { ResourceService } from './resource.service';
import { CartDialog } from './main/cart.component';
import {DragDropModule} from '@angular/cdk/drag-drop';
import { StaticlinksComponent } from './staticlinks/staticlinks.component';
import { CartComponent } from './cart/cart.component';
import { CheckoutComponent } from './checkout/checkout.component';
import { ListComponent } from './list/list.component';
import { ProductComponent } from './product/product.component';
import { CustomerComponent, AppContact } from './customer/customer.component';
import { StaticpageComponent } from './staticpage/staticpage.component'; 
import { AngularFontAwesomeModule } from 'angular-font-awesome';
import { E404Component } from './e404/e404.component';
import { SearchComponent } from './search/search.component';
import { CompleteComponent } from './complete/complete.component';

@NgModule({
  declarations: [
    AppComponent,
    ToolbarComponent,
    MainComponent,
    CartDialog,
    StaticlinksComponent,
    CartComponent,
    CheckoutComponent,
    ListComponent,
    ProductComponent,
    CustomerComponent,
    AppContact,
    StaticpageComponent,
    E404Component,
    SearchComponent,
    CompleteComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    MorphModule,
    MaterialModule,
    BrowserAnimationsModule,
    OverlayModule,
    AdminModule,
    DragDropModule,
    MDBBootstrapModule,
    AngularFontAwesomeModule
  ],
  providers: [ResourceService],
  entryComponents:[CartDialog,AppContact],
  bootstrap: [AppComponent]
})
export class AppModule { }
