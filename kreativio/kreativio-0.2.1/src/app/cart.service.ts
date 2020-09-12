import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class CartService {
  cart_storage_key = 'shopping_cart';
  parse_cart(){
    try{
    return JSON.parse(sessionStorage.getItem(this.cart_storage_key));
    }catch(e){console.log(e)} 
  }
  save_cart(cart){
    sessionStorage.setItem(this.cart_storage_key,JSON.stringify(cart));
  }
  init_cart() {
    sessionStorage.setItem(this.cart_storage_key,JSON.stringify({data:[]}))
  }
  add_to_cart(product){
    try{
    let cart = this.parse_cart();
    let inserted = false;
    cart.data.forEach(e=>{
      if(e.id==product.id&&!inserted) {e = product;inserted=true}
      else if(!inserted){cart.data.push(product);inserted=true}
    });
    if(cart.data.length==0||!inserted) cart.data.push(product)
    this.save_cart(cart);
    }catch(e){
      console.log('Failed to save items to cart.',e)
      return false
    }
  }
  remove_from_cart(id){
    let cart = this.parse_cart()
    let new_cart=[];
    cart.forEach(product=>{
      if(product.id!=id){
        new_cart.push(product);
      }
    });
    this.save_cart(new_cart);
  }
  remove_cart(){

  }

  constructor() { }
}
