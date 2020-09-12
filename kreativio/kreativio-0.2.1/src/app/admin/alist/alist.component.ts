import { Component, OnInit, OnDestroy, ViewChild } from '@angular/core';
import { ResourceService, ArticleID, CategoryListConditions, ListConditions } from 'src/app/resource.service';
import { ImageService } from 'schema-admin';
import { Subject } from 'rxjs';
import { LanguageService } from 'src/app/services/language.service';
import { takeUntil } from 'rxjs/operators';
import { FormControl } from '@angular/forms';
import { MatSlideToggle, MatSelect } from '@angular/material';

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.css']
})
export class AListComponent implements OnInit, OnDestroy {
  list;
  public language;
  public currency:string = 'Lei';
  private language_all;
  private u$:Subject<void> = new Subject<void>();
  constructor(private rs:ResourceService,private is:ImageService,private ls:LanguageService) {
    is.image_client_ept = "https://kreativio.ro:15002"
    this.rs.article_list.pipe(takeUntil(this.u$)).subscribe((data:any)=>{
      this.list=data;
    });
    ls.getLanguageLabels();
    this.ls.language.pipe(takeUntil(this.u$)).subscribe((data:any)=>{
      this.language_all = data;
      this.language = data.ro;
    });
    this.getCategoryList()
   }
   refreshList(){
     this.list_conditions = new ListConditions(); // reset list_conditions to default.
     this.categories_select.writeValue(null);
     this.only_promoted.writeValue(false);
     setTimeout(()=>{this.rs.listArticles();},300); // TODO - make something to prevent spam click
   }
   getLinksOfArticle(arr){
    let o = [];
    arr.forEach(e=>{
      o.push({label:e.getLabel(),url:e.getUrl(),id:e.getId()})
    });
    return o;
   }
  deleteArticle(id){
    let rq = new ArticleID();
    let shop_delete_success:boolean = false;
    rq.setToken(sessionStorage.getItem("t"));
    rq.setId(parseInt(id,10));
    this.rs.shopClient.deleteArticle(rq,{},(e,r)=>{
      if(!e){
        shop_delete_success = true;
        this.rs.notify('Articol sters.',3000);
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
    if(shop_delete_success){    
    this.list.forEach(item=>{
      let arr = [];
      if(item.id!=id){
        arr.push(id);
      }else if(item.id==id){
        item.images.forEach(i=>{
          this.DeleteFile(i.url);
        });
        item.videos.forEach(i=>{
          this.DeleteFile(i.url);
        });
      }
      this.list=arr;
    });
    }
  }
  DeleteFile(link:string){
    this.is.RemoveFile([link]).then(()=>{});
  }

  public categories;
  public category_selected = new FormControl();
  getCategoryList() {
    let c = new CategoryListConditions();
    this.rs.listCategories(c,(r)=>{
      this.categories = r.getListList();
    })
  }
  public list_conditions = new ListConditions();
  @ViewChild("onlyPromoted",{static:false,read:MatSlideToggle}) public only_promoted:MatSlideToggle;
  @ViewChild("categoriesSelect",{static:false,read:MatSelect}) public categories_select:MatSelect; 
  categorySelected(event) {
    console.log(this.list_conditions);
    if(this.only_promoted.checked) this.list_conditions.setOnlyPromoted(true);
    else this.list_conditions.setOnlyPromoted(false);
    if(event.value) this.list_conditions.setOnlyCategoryId(event.value.getId());
    this.rs.shopClient.listArticles(this.list_conditions,{},(e,r)=>{
      if(!e){
        this.list = [];
        r.getListList().forEach((i:any)=>{
          this.list.push({
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
    this.rs.listArticles();
  }
  ngOnDestroy() {
    this.u$.next();
    this.u$.unsubscribe();
  }

}
