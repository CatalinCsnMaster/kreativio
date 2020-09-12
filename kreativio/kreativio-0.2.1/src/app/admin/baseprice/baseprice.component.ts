import { Component, OnInit } from '@angular/core';
import { ResourceService, BasePrice, BasePriceListCondtions } from 'src/app/resource.service';

@Component({
  selector: 'app-baseprice',
  templateUrl: './baseprice.component.html',
  styleUrls: ['./baseprice.component.css']
})
export class BasepriceComponent implements OnInit {

  constructor(private rs:ResourceService) {
    this.listBasePrices();
   }
  baseprices:Array<any> = [];

  listBasePrices() {
    let lc = new BasePriceListCondtions();
    this.rs.shopClient.listBasesPrices(lc,{},(e,r)=>{
      if(!e) {
        this.baseprices = r.getListList();
        console.log(r);
      } else{
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

  addBasePrice(label,price) {
    let bp = new BasePrice();
    bp.setLabel(label.value);
    bp.setPrice(price.value);
    bp.setToken(localStorage.getItem("t"));
    this.rs.shopClient.saveBasePrice(bp,{},(e,r)=>{
      if(!e) { 
        this.baseprices.push(r);
        this.rs.notify("Adaugat",3000);
      } else{
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
  removeBasePrice(label) {
    let arr = [];
    this.baseprices.forEach(e=>{
      if(e.getLabel()!=label) arr.push(e);
      else {
        e.setToken(localStorage.getItem("t"));
        this.rs.shopClient.deleteBasePrice(e,{},(e,r)=>{
          if(!e) {
            this.rs.notify("Sters",3000);
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
        })
      }
    });
    this.baseprices = arr;
  }

  ngOnInit() {
  }

}
