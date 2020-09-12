import { Component, OnInit, OnDestroy, ViewChild, AfterContentInit, ElementRef } from '@angular/core';
import { CartService } from '../cart.service';
import { ResourceService, Order,ArticleAmount, OrderID } from '../resource.service';
import { FormGroup, FormControl, Validators } from '@angular/forms';
import { LanguageService } from '../services/language.service';
import { takeUntil } from 'rxjs/operators';
import { Subject } from 'rxjs';
import { CallService } from 'schema-admin';

@Component({
  selector: 'app-checkout',
  templateUrl: './checkout.component.html',
  styleUrls: ['./checkout.component.css']
})
export class CheckoutComponent implements OnInit, OnDestroy, AfterContentInit {
  public data:Array<any> = [
    {product:"Article 1",quantity:1,price: 199.90},
    {product:"Article 2",quantity:1,price: 399.90},
    {product:"Article 3",quantity:1,price: 2399.90},
    {product:"Article 4",quantity:1,price: 499.90},
    {product:"Article 5",quantity:1,price: 799.90}
  ];
  subtotal:number=0;
  delivery:number=0;
  total:number=0;
  currency:string='LEI';
  payment_choice;
  language;
  private u$:Subject<void> = new Subject<void>();
  constructor(private cart:CartService,private rs:ResourceService,public ls:LanguageService,public call:CallService) { 
    this.data = cart.parse_cart().data;
    this.data.forEach(product=>{this.subtotal+=product.price_t});
    if(this.payment_choice==0) this.total=this.subtotal+this.delivery;
    else this.total=this.subtotal;
    this.ls.getLanguageLabels();
    this.ls.language.pipe(takeUntil(this.u$)).subscribe(data=>this.language=data.ro);
    this.rs.getStaticSchema("https://kreativio.ro:15001",call,"kreativio-ro");
    this.call.schema.pipe(takeUntil(this.u$)).subscribe((data)=>{
      this.delivery = parseFloat(data.read.livrare.value);
    });
  }
  log(v){
    console.log(v)
  }
  mobilpay_redirect:boolean = false;
  recalc(){
    this.subtotal = 0;
    this.total = 0;
    this.data = this.cart.parse_cart().data;
    this.data.forEach(product=>{this.subtotal+=product.price_t});
    if(this.payment_choice==0) this.total=this.subtotal+this.delivery;
    else this.total = this.subtotal;
    if(this.payment_choice==2) {
      this.mobilpay_redirect = true
    } else this.mobilpay_redirect = false;
  }
  orderinfo = new FormGroup({
    fullname:new FormControl('', {validators: Validators.minLength(3)}),
    email:new FormControl('', {validators: Validators.minLength(7)}),
    phone:new FormControl('', {validators: Validators.minLength(6)}),
    address:new FormControl('', {validators: Validators.minLength(10)}),
    message:new FormControl('', {validators: Validators.required}),
  })
  @ViewChild("redirect",{read:HTMLFormElement,static:true}) public payment_redirect_form:HTMLFormElement;
  @ViewChild("formWrapper",{static:false}) formWrapper;
  public encKey:string;
  public encData:string;
  checkout(redirect?:boolean){
    let o = new Order();
    let articles = [];
    this.data.forEach(article=>{
      let a = o.addArticles();
      a.setAmount(article.quantity);
      a.setArticleId(article.id);
      if(article.variant_id!=undefined) a.setVariantId(article.variant_id);
      if(article.base_price_id!=undefined) a.setBasePriceId(article.base_price_id);
      articles.push(a);
    });
    o.setArticlesList(articles);
    o.setFullName(this.orderinfo.get('fullname').value);
    o.setFullAddress(this.orderinfo.get('address').value);
    o.setMessage(this.orderinfo.get('message').value);
    o.setEmail(this.orderinfo.get('email').value);
    o.setPhone(this.orderinfo.get('phone').value);
    o.setPaymentMethod(parseInt(this.payment_choice,10));
    o.setStatus(0);

    this.rs.shopClient.checkout(o,{},(e,r)=>{
      if(!e){
        if(!redirect){
          this.rs.notify(this.language.order_sent,9000);
          setTimeout(()=> {localStorage.clear();sessionStorage.clear();location.replace('/')},9000);
        }else{
          localStorage.clear();
          sessionStorage.clear();
          this.encKey = r.getEnvKey();
          this.encData = r.getData();
          setTimeout(()=>this.formWrapper.nativeElement.childNodes[0].submit(),1000);
        }
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

  ngAfterContentInit() {
  }

  ngOnDestroy() {
    this.u$.next();
    this.u$.unsubscribe();
  }
}
