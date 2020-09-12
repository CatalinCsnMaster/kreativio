import { Component, OnInit } from '@angular/core';
import { ResourceService, ListOrderConditions } from 'src/app/resource.service';
import { MatDialog, MatBottomSheet } from '@angular/material';
import { OrderComponent } from '../order/order.component';

@Component({
  selector: 'app-orders',
  templateUrl: './orders.component.html',
  styleUrls: ['./orders.component.css'],
  host: {
    class: 'light-mode bg-light2'
  }
})
export class OrdersComponent implements OnInit {

  constructor(private rs:ResourceService, private dialog:MatBottomSheet) {
    this.order_conditions.setToken(sessionStorage.getItem("t"));
    this.order_conditions.setStatus(0);
    this.getOrders();
  }
  
  order_conditions = new ListOrderConditions();
  orders;
  currency = 'LEI';

  listChange(event){
    this.order_conditions.setToken(sessionStorage.getItem("t"));
    this.order_conditions.setStatus(event.index);
    this.getOrders();
  }

  getOrders(){
    this.order_conditions.setToken(sessionStorage.getItem("t"));
    this.rs.shopClient.listOrders(this.order_conditions,{},(e,r)=>{
      if(!e){
        this.orders = r.getListList();
      }else{
        this.rs.errorSwitch(e);
        if(e.code==16) {
        setTimeout(()=>{
          localStorage.clear();
          sessionStorage.clear();
          location.replace('/admin');
        },3000);
        }
      }
    });
  }

  tabs = [
    {label:"Toate comenzile",icon:"list-alt",content:[]},
    {label:"Comenzi deschise",icon:"turned_in_not",content:[]},
    {label:"Comenzi trimise",icon:"mail_outline",content:[]},
    {label:"Comenzi inchise",icon:"turned_in",content:[]},
  ];

  detail(order){
    this.dialog.open(OrderComponent,{data:order,panelClass:'light-mode'}).afterDismissed().toPromise().then(()=>{this.getOrders()});
  }

  ngOnInit() {
  }

}
