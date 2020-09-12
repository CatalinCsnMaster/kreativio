import { Component, OnInit, AfterContentInit, Input, Inject } from '@angular/core';
import { ResourceService } from '../resource.service';
import { Subject } from 'rxjs';
import { Platform } from '@angular/cdk/platform';
import { MatDialog } from '@angular/material';
import { CartDialog } from './cart.component';
import { takeUntil } from 'rxjs/operators';
import { PlatformService } from '../services/platform.service';
import { CallService } from 'schema-admin';
import { inject } from '@angular/core/testing';
import { DomSanitizer } from '@angular/platform-browser';

@Component({
  selector: 'app-main',
  templateUrl: './main.component.html',
  styleUrls: ['./main.component.css']
})
export class MainComponent implements OnInit,AfterContentInit {
  promo_data;
  private u$:Subject<any> = new Subject();
  carousel_images = [];
  about_paragraph;
  constructor(
    private resource_service:ResourceService,
    public platform:PlatformService,
    public call:CallService,
    private sanitizer:DomSanitizer) 
   {  this.resource_service.listArticles(true,true);
      this.resource_service.article_list.pipe(takeUntil(this.u$)).subscribe(data=>{
      this.promo_data=data});
      this.resource_service.getStaticSchema("https://kreativio.ro:15001",call,"kreativio-ro");
      this.call.schema.pipe(takeUntil(this.u$)).subscribe((data)=>{
        this.carousel_images = data.read.carousel.value;
        this.about_paragraph = data.read.despre.value;
      });
   }
   cleanLink(v){
     return this.sanitizer.bypassSecurityTrustUrl(v);
   }
 ngOnInit() {
  }
  ngAfterContentInit(): void {
  }
  ngOnDestroy(): void {
    //Called once, before the instance is destroyed.
    //Add 'implements OnDestroy' to the class.
    this.u$.next();
    this.u$.unsubscribe();
  }
}
