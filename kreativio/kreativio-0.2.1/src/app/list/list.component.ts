import { Component, OnInit, OnDestroy } from '@angular/core';
import { ResourceService, ListConditions } from '../resource.service';
import { PlatformService } from '../services/platform.service';
import { ActivatedRoute, Params } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil, take } from 'rxjs/operators';

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.css']
})
export class ListComponent implements OnInit, OnDestroy {
  public data:any;
  private u$:Subject<void> = new Subject<void>();
  constructor(private resource_service:ResourceService,public platform:PlatformService,public activeRoute:ActivatedRoute) {
    
    this.activeRoute.queryParams.pipe(takeUntil(this.u$)).subscribe((paramMap:Params)=>{
        let category_label = paramMap['category'];
        if(category_label) this.getArticles(category_label);
        else {
          this.resource_service.listArticles(true,false);
          this.resource_service.article_list.pipe(takeUntil(this.u$)).subscribe(data=>{
            this.data=data});
        }
      });

  }
  getArticles(category:string) {
    let c = new ListConditions();
    c.setOnlyCategoryLabel(category);
    c.setOnlyPublished(true);
    this.resource_service.shopClient.listArticles(c,{},(e,r)=>{
      if(!e){
        this.data = [];
        r.getListList().forEach((i:any)=>{
          this.data.push({
            id:i.getId(),
            title:i.getTitle(),
            description:i.getDescription(),
            price:i.getPrice(),
            published:i.getPublished(),
            promoted:i.getPromoted(),
            images:this.resource_service.getLinksOfArticle(i.getImagesList()),
            videos:this.resource_service.getLinksOfArticle(i.getVideosList())
          });
        });
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
    });
  }

  ngOnInit() {
  }
  ngOnDestroy(): void {
    this.u$.next();
    this.u$.unsubscribe();
  }
}
