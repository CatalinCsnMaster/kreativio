import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { MainComponent } from './main/main.component';
import { AdminComponent } from './admin/admin/admin.component';
import { AuthComponent, AuthGuard, MorphComponent } from 'schema-admin';
import { CheckoutComponent } from './checkout/checkout.component';
import { CartComponent } from './cart/cart.component';
import { ProductComponent } from './product/product.component';
import { ListComponent } from './list/list.component';
import { AListComponent } from './admin/alist/alist.component'
import { CustomerComponent } from './customer/customer.component';
import { StaticpageComponent } from './staticpage/staticpage.component';
import { ArticleComponent } from './admin/article/article.component';
import { OrdersComponent } from './admin/orders/orders.component';
import { OrderComponent } from './admin/order/order.component';
import { CategoriesComponent } from './admin/categories/categories.component';
import { E404Component } from './e404/e404.component';
import { SearchComponent } from './search/search.component';
import { BasepriceComponent } from './admin/baseprice/baseprice.component';
import { CompleteComponent } from './complete/complete.component';

const adminRoutes:Routes = [
  {path:"article",component:ArticleComponent},
  {path:"article:id",component:ArticleComponent}, 
  {path:"kreativio-ro", component:MorphComponent},
  {path:"list",component:AListComponent},
  {path:"orders",component:OrdersComponent},
  {path:"categories",component:CategoriesComponent},
  {path:"materiale",component:BasepriceComponent}
];
const customerRoutes:Routes = [
  {path:"",component:MainComponent},
  {path:"checkout",component:CheckoutComponent},
  {path:"cart",component:CartComponent},
  {path:"product",component:ProductComponent},
  {path:"product:id",component:ProductComponent},
  {path:"list",component:ListComponent},
  {path:"list:category",component:ListComponent},
  {path:"static",component:StaticpageComponent},
  {path:"static:view",component:StaticpageComponent},
  {path:"search",component:SearchComponent},
  {path:"search:q",component:SearchComponent},
  {path:"sent",component:CompleteComponent},
  {path:"send:id",component:CompleteComponent}
];
const routes: Routes = [
  {path:"",component:CustomerComponent,children:customerRoutes},
  {path:"admin",component:AdminComponent, children:adminRoutes, canActivate:[AuthGuard]},
  {path:"auth",component:AuthComponent},
  {path:"auth:token",component:AuthComponent},
  {path:"**",component:E404Component}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
