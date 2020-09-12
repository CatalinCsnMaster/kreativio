import { Component, OnInit, OnDestroy, ViewChild, AfterContentInit, AfterViewInit } from '@angular/core';
import { FormGroup, FormControl, Validators } from '@angular/forms';
import { Article, Media, ResourceService, ArticleID, CategoryListConditions, Variant, BasePriceListCondtions } from 'src/app/resource.service';
import { ImageService } from 'schema-admin';
import { Subject } from 'rxjs';
import { ActivatedRoute, Router } from '@angular/router';
import { moveItemInArray, CdkDragDrop } from '@angular/cdk/drag-drop';
import { LanguageService } from 'src/app/services/language.service';
import { takeUntil } from 'rxjs/operators';
import { MatOptgroup, MatOption, MatSelect } from '@angular/material';

const { ImageServiceClient } = require('proto/image_api_grpc_web_pb.js');
const { NewImageRequest, NewImageResponse, RemoveImageRequest, RemoveImageResponse } = require('proto/image_api_pb.js');

@Component({
  selector: 'app-article',
  templateUrl: './article.component.html',
  styleUrls: ['./article.component.css']
})
export class ArticleComponent implements OnInit, OnDestroy {
  public images = [];
  public videos = [];
  public id:number = 0;
  article = new FormGroup({
    title: new FormControl('',Validators.required),
    description: new FormControl('',Validators.required),
    price: new FormControl(0,Validators.required),
    publish: new FormControl(),
    promoted: new FormControl()
  });
  public language;
  public variants;
  public selected_base_prices;
  public base_prices;
  private language_all;
  private u$:Subject<void> = new Subject<void>();

  constructor(private is:ImageService,private rs:ResourceService,private ar:ActivatedRoute,private router:Router, private ls:LanguageService) {
    is.image_client_ept = "https://kreativio.ro:15002"
    this.ar.queryParamMap.pipe(takeUntil(this.u$)).subscribe(paramMap=>{
      this.id = parseInt(paramMap.get('id'),10);
      if(this.id!=0) {
        this.getArticle(paramMap.get('id'));
      } else { // if url changes while component is already loaded, reallocate defaults.
        this.article = new FormGroup({
          title: new FormControl('',Validators.required),
          description: new FormControl('',Validators.required),
          price: new FormControl('',Validators.required),
          publish: new FormControl(),
          promoted: new FormControl()
        });
        this.images = [];
        this.videos = [];
      }
    });
    ls.getLanguageLabels();
    this.ls.language.pipe(takeUntil(this.u$)).subscribe((data:any)=>{
      this.language_all = data;
      this.language = data.ro;
    });
    this.getCategoryList();
    this.listBasePrices();
   }
  getLinksOfArticle(arr){
    let o = [];
    arr.forEach(e=>{
      o.push({label:e.getLabel(),url:e.getUrl(),id:e.getId()})
    });
    return o;
  }
  getVariants(arr) {
    let o = [];
    arr.forEach(e=>{
      o.push({labels:e.getLabelsList(),multiplier:e.getMultiplier(),id:e.getId()})
    });
    return o;
  }
  addVariant(label1,label2,multiplier) {
    if(this.variants==undefined) this.variants = [];
    var lbls = [label1.value.trim(),label2.value.trim()];
    this.variants.push({labels:lbls,multiplier:multiplier.value});
  }
  removeVariant(i) {
    let vr = [];
    for(let y=0;y<this.variants.length;y++) {
      if(i!=y) {
        vr.push(this.variants[y]);
      }
    }
    this.variants = vr;
  }
  public view_article:any = new Object();
  getArticle(id){
    let rq = new ArticleID();
    rq.setId(id)
    this.rs.shopClient.viewArticle(rq,{},(e,r)=>{
      if(!e){
        this.view_article.title = r.getTitle();
        this.view_article.description = r.getDescription();
        this.view_article.price = r.getPrice();
        this.view_article.promoted = r.getPromoted();
        this.view_article.published = r.getPublished();
        this.images = this.getLinksOfArticle(r.getImagesList());
        this.videos = this.getLinksOfArticle(r.getVideosList());
        this.variants = this.getVariants(r.getVariantsList());
        this.selected_base_prices = r.getBasepricesList();
        this.article.setValue({
          promoted:r.getPromoted(),
          publish:r.getPublished(),
          title:r.getTitle(),
          description:r.getDescription(),
          price: parseFloat(r.getPrice())
        });
        this.categories_selected_value = r.getCategoriesList();
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
  afterUploadImage(response,cls,label){
    cls.images.push({url:response.getLinkList()[0],label:label});
  }
  afterUploadVideo(response,cls,label){
    cls.videos.push({url:response.getLinkList()[0],label:label});
  }
  UploadImage(f,label){
    this.uploadFile([f.files[0]],this.afterUploadImage,label);
  }
  UploadVideo(f,label){;
    this.uploadFile([f.files[0]],this.afterUploadVideo,label);
  }
  async NewImage(f: File[],cb:any,label) {
    let images_client = new ImageServiceClient('https://kreativio.ro:15002')
    let rqst = new NewImageRequest();
    rqst.setTkn(localStorage.getItem('t'));
    let files:Array<Uint8Array> = [];
    let fnc = async ()=>{
    for(let i=0; i<f.length; i++){
      let data = await new Response(f[i]).arrayBuffer();
      let bytes = new Uint8Array(data);
      files.push(bytes);
    };}
    return await fnc().then(()=>{rqst.setImageList(files)}).then(()=>{
      images_client.newImagePreserve(rqst,{}, (e,rsp)=>{
        if(!e){
          return cb(rsp,this,label);
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
    });
  }
  uploadFile(f:File[],cb,label){
    this.NewImage(f,cb,label);
  }
  DeleteFile(tp,link:string){
    this.is.RemoveFile([link]).then(()=>{
      let _copy = [];
      if(tp=="image"){
      this.images.forEach(item=>{
        if(item.url!=link) _copy.push(item);
      });
      this.images = _copy;
      }else if(tp=="video"){
        this.videos.forEach(item=>{
          if(item.url!=link) _copy.push(item);
        });
        this.videos = _copy;
      }
    });
  }
  save(){
    let a = new Article();
    a.setId(this.id);
    a.setTitle(this.article.get('title').value);
    a.setDescription(this.article.get('description').value);
    a.setPrice(this.article.get('price').value.toString());
    this.images.forEach(i=>{
      let image = new Media();
      image.setLabel(i.label);
      image.setUrl(i.url);
      a.addImages(image);
    });
    this.videos.forEach(v=>{
      let video = new Media();
      video.setLabel(v.label);
      video.setUrl(v.url);
      a.addVideos(video);
    });
    
    if(this.variants&&this.variants.length!=0) this.variants.forEach(v=>{
      let vr = new Variant();
      vr.setLabelsList(v.labels);
      vr.setMultiplier(v.multiplier);
      vr.setId(v.id);
      a.addVariants(vr);
    });
    a.setToken(sessionStorage.getItem('t'));
    a.setPublished(this.article.get('publish').value);
    a.setPromoted(this.article.get('promoted').value);
    a.setCategoriesList(this.categories_selected_value);
    a.setBasepricesList(this.selected_base_prices);
    this.rs.saveArticle(a);
    this.rs.notify('Salvat',9000);
    this.router.navigate(['/admin/list']);
  }
  dropped(event){
    moveItemInArray(event.container.data,event.previousIndex,event.currentIndex)
  }
  public categories;
  public categories_selected_value = [];
  getCategoryList() {
    let c = new CategoryListConditions();
    this.rs.listCategories(c,(r)=>{
      this.categories = r.getListList();
    })
  }
  listBasePrices() {
    let lc = new BasePriceListCondtions();
    this.rs.shopClient.listBasesPrices(lc,{},(e,r)=>{
      if(!e) {
        this.base_prices = r.getListList();
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
  ngOnDestroy(): void {
    this.u$.next();
    this.u$.unsubscribe();
  }
  log(event){
    console.log(event);
  }
  ngOnInit() {
  }
  compareCategories(i1,i2){
    return i1 && i2 ? i1.getLabel() === i2.getLabel() : i1 === i2;
  }
  compareBasePrices(i1,i2) {
    return i1 && i2 ? i1.getId() === i2.getId() : i1 === i2;
  }
}
