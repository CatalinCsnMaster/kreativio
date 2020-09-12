import { Component, OnInit, Inject } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA, MatBottomSheet, MAT_BOTTOM_SHEET_DATA, MatBottomSheetRef } from '@angular/material';
import { ResourceService } from 'src/app/resource.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-order',
  templateUrl: './order.component.html',
  styleUrls: ['./order.component.css'],
  host: {
    class: 'light-mode bg-light2',
  }
})
export class OrderComponent implements OnInit {

  constructor(private dialogRef:MatBottomSheetRef<OrderComponent>,@Inject(MAT_BOTTOM_SHEET_DATA) public order,private rs:ResourceService,private router:Router) {
  }
  state;
  currency = "Lei";
  saveOrderWith(){
    console.log(this.order);
    this.order.setStatus(this.state);
    this.order.setToken(sessionStorage.getItem("t"));
    this.rs.shopClient.saveOrder(this.order,{},(e,r)=>{
      if(!e){
        console.log(r);
        this.rs.notify('Modificari salvate.',9000);
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

  ngOnInit() {
  }

}
