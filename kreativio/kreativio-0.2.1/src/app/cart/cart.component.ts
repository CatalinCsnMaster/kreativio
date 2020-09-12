import { Component, OnInit, IterableDiffers, OnDestroy } from '@angular/core';
import { ResourceService } from '../resource.service';
import { CartService } from '../cart.service';
import { LanguageService } from '../services/language.service';
import { takeUntil } from 'rxjs/operators';
import { Subject } from 'rxjs';

@Component({
  selector: 'app-cart',
  templateUrl: './cart.component.html',
  styleUrls: ['./cart.component.css']
})
export class CartComponent implements OnInit, OnDestroy {
  public data:Array<any> = [];
  public language;
  private u$:Subject<void> = new Subject<void>();
  minus(q:number){
   if(this.data[q].quantity>1){
     this.data[q].quantity -=1
     this.data[q].price_t -= parseFloat(this.data[q].price);
    }
   this.cart.save_cart({data:this.data});
  }
  plus(q:number){
    this.data[q].quantity +=1;
    this.data[q].price_t += parseFloat(this.data[q].price);
    this.cart.save_cart({data:this.data});
  }
  remove(i:number){
    let dt = [];
    for(let z=0;z<this.data.length;z++){
      if(z!=i) dt.push(this.data[z]);
    }
    this.data = dt;
    this.cart.save_cart({data:this.data});
  }
  empty(){
    this.cart.save_cart({data:[]});
    this.data = [];
    this.resource_service.notify('Cart emptied.',300);
  }
  constructor(public resource_service:ResourceService, private cart:CartService, public ls:LanguageService) { 
    this.data = cart.parse_cart().data;
    this.ls.getLanguageLabels();
    this.ls.language.pipe(takeUntil(this.u$)).subscribe(data=>this.language=data.ro);
  }


  ngOnInit() {
  }
  ngOnDestroy(): void {
    //Called once, before the instance is destroyed.
    //Add 'implements OnDestroy' to the class.
    this.u$.next();
    this.u$.unsubscribe();
  }
}
