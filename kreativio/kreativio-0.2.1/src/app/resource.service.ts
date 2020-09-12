import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Subject } from 'rxjs';
import { MatSnackBar } from '@angular/material';
import { CallService } from 'schema-admin';

const { Article, ArticleID, ListConditions, Media, ArticleList,Order,PaymentMethod,ArticleAmount, ListOrderConditions, CategoryList, CategoryListConditions, Category, TextSearch,SuggestionList, OrderID, BasePrice, Variant, Details, BasePriceListCondtions, Message, MessageID } = require('proto/shop_pb.js');
const { ShopClient } = require('proto/shop_grpc_web_pb.js');
export { Article,Media,ArticleID,ListConditions,Order,ArticleList,PaymentMethod,ArticleAmount,ListOrderConditions, CategoryList, CategoryListConditions, Category, TextSearch, SuggestionList, OrderID, BasePrice, Variant, Details, BasePriceListCondtions, Message, MessageID, ShopClient };
@Injectable({
  providedIn: 'root'
})
export class ResourceService {
public promo_data:Subject<any> = new Subject();
public shopClient;
public shopServiceEndpoint = 'https://kreativio.ro:15003';
  constructor(private http:HttpClient,private snb:MatSnackBar) { 
    this.getResource();
    this.shopClient = new ShopClient(this.shopServiceEndpoint);
  }
  public article_list:Subject<any>=new Subject<any>();
  public listArticles(published?,promoted?){
    let c = new ListConditions();
    if(published)c.setOnlyPublished(true);
    if(promoted)c.setOnlyPromoted(true);
    this.shopClient.listArticles(c,{},(e,r)=>{
      if(!e){
        let list=[];
        r.getListList().forEach((i:any)=>{
          list.push({
            id:i.getId(),
            title:i.getTitle(),
            description:i.getDescription(),
            price:i.getPrice(),
            published:i.getPublished(),
            promoted:i.getPromoted(),
            images:this.getLinksOfArticle(i.getImagesList()),
            videos:this.getLinksOfArticle(i.getVideosList())
          });
        });
        this.article_list.next(list);
      }else{
        this.errorSwitch(e);
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
  getLinksOfArticle(arr){
    let o = [];
    arr.forEach(e=>{
      o.push({label:e.getLabel(),url:e.getUrl(),id:e.getId()})
    });
    return o;
   }
  public viewArticle(id:number){
    let rq = new ArticleID();
    console.log(rq);
    rq.setId(id)
    this.shopClient.viewArticle(rq,{},(e,r)=>{
      if(!e){
        return r
      }else{
        this.errorSwitch(e);
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

  public errorSwitch(e:any) {
    switch(e.code){
      case 1:
        this.notify(`<CODE ${e.code} - CANCELLED> `+e.message,9000);
      break;
      case 2:
        this.notify(`<CODE ${e.code} - UNKNOWN> `+e.message,9000);
      break;
      case 3:
        this.notify(`<CODE ${e.code} - INVALID_ARGUMENT> `+e.message,9000);
      break;
      case 4:
        this.notify(`<CODE ${e.code} - DEADLINE_EXCEEDED> `+e.message,9000);
      break;
      case 5:
        this.notify(`<CODE ${e.code} - NOT_FOUND> `+e.message,9000);
      break;
      case 6:
        this.notify(`<CODE ${e.code} - ALREADY_EXISTS> `+e.message,9000);
      break;
      case 7:
        this.notify(`<CODE ${e.code} - PERMISSION_DENIED> `+e.message,9000);
      break;
      case 16:
        this.notify(`<CODE ${e.code} - UNAUTHENTICATED> `+e.message,9000);
      break;
      case 8:
        this.notify(`<CODE ${e.code} - RESOURCE_EXHAUSTED> `+e.message,9000);
      break;
      case 9:
        this.notify(`<CODE ${e.code} - FAILED_PRECONDITION> `+e.message,9000);
      break;
      case 10:
        this.notify(`<CODE ${e.code} - ABORTED> `+e.message,9000);
      break;
      case 11:
        this.notify(`<CODE ${e.code} - OUT_OF_RANGE> `+e.message,9000);
      break;
      case 12:
        this.notify(`<CODE ${e.code} - UNIMPLEMENTED> `+e.message,9000);
      break;
      case 13:
        this.notify(`<CODE ${e.code} - INTERNAL> `+e.message,9000);
      break;
      case 14:
        this.notify(`<CODE ${e.code} - UNAVAILABLE> `+e.message,9000);
      break;
      case 15:
        this.notify(`<CODE ${e.code} - DATA_LOSS> `+e.message,9000);
      break;
      default:
        this.notify(e.message,9000);
    }
  }

  public saveArticle(a){
    this.shopClient.saveArticle(a,{},(e,r)=>{
      if(!e){
        return r;
      }else{
        this.errorSwitch(e);
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
  public getResource(cb?){
    this.http.get('/assets/fake_resource.json').toPromise().then((data)=>{
      this.promo_data.next(data);
      if(cb)cb();
    });
  }
  public getStaticSchema(ept:string,service:CallService,key:string){
    service.schema_ept_1 = ept;
    service.getSchema(key);
  }
  public notify(msg:string,d:number){
    this.snb.open(`🔔 ${msg}`,'❌',{duration:d});
  }
  public listCategories(categoryListConditions,cb?) {
    this.shopClient.listCategories(categoryListConditions,{},(e,r)=>{
      if(!e){
        if(cb)cb(r);
      }else{
        this.errorSwitch(e);
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
}
