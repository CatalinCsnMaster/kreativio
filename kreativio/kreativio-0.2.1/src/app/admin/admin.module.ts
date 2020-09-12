import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AdminComponent } from './admin/admin.component';
import { MorphModule } from 'schema-admin';
import { RouterModule } from '@angular/router';
import { MaterialModule } from '../material/material.module';
import { ArticleComponent } from './article/article.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { AListComponent } from './alist/alist.component';
import { OrdersComponent } from './orders/orders.component';
import { OrderComponent } from './order/order.component';
import { DragDropModule } from '@angular/cdk/drag-drop';
import { CategoriesComponent, CatEdit } from './categories/categories.component';
import { BasepriceComponent } from './baseprice/baseprice.component';



@NgModule({
  declarations: [AdminComponent, ArticleComponent, AListComponent, OrdersComponent, OrderComponent, CategoriesComponent, BasepriceComponent, CatEdit],
  imports: [
    CommonModule,
    MaterialModule,
    FormsModule,
    ReactiveFormsModule,
    MorphModule,
    RouterModule,
    DragDropModule
  ],
  entryComponents: [ OrderComponent, CatEdit ]
})
export class AdminModule { }
