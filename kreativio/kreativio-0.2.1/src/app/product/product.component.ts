import { Component, OnInit, OnDestroy } from '@angular/core';
import { ResourceService, ArticleID } from '../resource.service';
import { takeUntil } from 'rxjs/operators';
import { Subject } from 'rxjs';
import { ActivatedRoute, ParamMap } from '@angular/router';
import { CartService } from '../cart.service';
import { parse } from 'querystring';

@Component({
  selector: 'app-product',
  templateUrl: './product.component.html',
  styleUrls: ['./product.component.css']
})
export class ProductComponent implements OnInit, OnDestroy {
  public data:any;
  public view_data:any = new Object();
  private u$:Subject<void> = new Subject<void>(); 
  constructor(private resource_service:ResourceService,
    private route:ActivatedRoute,
    private cart:CartService) {
      this.route.queryParamMap.pipe(takeUntil(this.u$)).subscribe((map:ParamMap)=>{
        let id = map.get('id');
        this.getArticle(id);
      });
  }
  variants;
  variants_parsed_list;
  sorted_widths;
  available_diameters;
  base_prices;
  base_price_selected;
  default_price;
  default_width;
  getArticle(id){
    let rq = new ArticleID();
    rq.setId(id)
    this.resource_service.shopClient.viewArticle(rq,{},(e,r)=>{
      if(!e){
        this.view_data.quantity = 1;
        this.view_data.id = r.getId();
        this.view_data.title = r.getTitle();
        this.view_data.description = r.getDescription();
        this.view_data.images = this.getLinksOfArticle(r.getImagesList());
        this.view_data.videos = this.getLinksOfArticle(r.getVideosList());
        this.view_data.promoted = r.getPromoted();
        this.variants = r.getVariantsList();
        this.base_prices = r.getBasepricesList();
        this.view_data.price = r.getPrice();
        this.view_data.price_t = parseFloat(r.getPrice());
        this.sortVariants();
        if(r.getBasepricesList().length!=0&&r.getVariantsList().length!=0){
          this.default_price = r.getBasepricesList()[0];
          this.base_price_selected = r.getBasepricesList()[0].getPrice();
          this.view_data.price = r.getBasepricesList()[0].getPrice(); // defaults
          this.view_data.price_t = r.getBasepricesList()[0].getPrice(); // defaults
          this.selected_width = this.sorted_widths[0].width; // defaults
          this.default_width = this.sorted_widths[0];
          this.available_diameters = this.sorted_widths[0].diameters;
          this.selected_diameter = this.sorted_widths[0].diameters[0]; // defaults
          for(let i=0;i<this.variants.length;i++) {
            if(this.variants[i].getLabelsList()[0] == this.selected_width && this.variants[i].getLabelsList()[1] == this.selected_diameter) {
              this.selected_multiplier = this.variants[i].getMultiplier();
              this.selected_variant_id = this.variants[i].getId();
            }
          }
          this.calc();
        }      
      }else{
        this.resource_service.errorSwitch(e);
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

  sortVariants() {
    let widths = new Object();
    this.sorted_widths = [];
    this.variants.forEach(v=>{
      widths[v.getLabelsList()[0].trim()] = [];
    });
    let width_keys = Object.keys(widths);
    for(let i = 0;i<this.variants.length;i++) {
        width_keys.forEach(w=>{
          if(this.variants[i].getLabelsList()[0].trim() == w) {
            widths[w].push(this.variants[i].getLabelsList()[1]);
          }
        });
    }
    for(let i=0;i<width_keys.length;i++) {
      this.sorted_widths.push({width:width_keys[i],diameters:widths[width_keys[i]]})
    }
  }
  calc() {
    this.view_data.price_t = this.base_price_selected * this.selected_multiplier;
    this.view_data.price = this.base_price_selected * this.selected_multiplier;
    this.view_data.quantity = 1;
  }
  base_price_id;
  selectBasePrice(ev) {
    this.base_price_selected = ev.value.getPrice();
    this.base_price_id = ev.value.getId();
    this.calc();
  }
  selected_width;
  selectDiameter(ev) {
    this.available_diameters = ev.value.diameters;
    this.selected_width = ev.value.width;
  }
  selected_diameter;
  selected_multiplier;
  selected_variant_id;
  estimatePrice(ev) {
    this.selected_diameter = ev.value;
    console.log(ev);
    for(let i=0;i<this.variants.length;i++) {
      if(this.variants[i].getLabelsList()[0] == this.selected_width && this.variants[i].getLabelsList()[1] == this.selected_diameter) {
          this.selected_multiplier = this.variants[i].getMultiplier();
          this.selected_variant_id = this.variants[i].getId();
          this.calc();
      }
    }
  }
  getLinksOfArticle(arr){
    let o = [];
    arr.forEach(e=>{
      o.push({label:e.getLabel(),url:e.getUrl(),id:e.getId()})
    });
    return o;
  }
  checkRoute(){
    this.route.queryParamMap.pipe(takeUntil(this.u$)).subscribe((map:ParamMap)=>{
      let id = map.get('id');
    });
  }

  minus(){
    if(this.view_data.quantity>1){
      this.view_data.quantity -=1;
      this.view_data.price_t -= parseFloat(this.view_data.price);
    }
  }
  plus(){
      this.view_data.quantity +=1;
      this.view_data.price_t += parseFloat(this.view_data.price);
  }
  add_to_cart(){
    if(this.variants.length!=0&&this.base_prices.length!=0){
      this.view_data['base_price_id'] = this.base_price_id;
      this.view_data['variant_id'] = this.selected_variant_id;
      if(this.selected_variant_id==undefined) {
        this.resource_service.notify('Adaugarea un cos a esuat, va rugam incercati mai tarziu.',9000);
        return;
      }
    }
    let o = this.cart.add_to_cart(this.view_data);
    if(o!=false){
      this.resource_service.notify('Adaugat in cos.',3000);
    }
  }
  ngOnInit() {
  }
  ngOnDestroy(){
    this.u$.next();
    this.u$.unsubscribe();
  }
}
