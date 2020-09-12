import { Component, OnInit, OnDestroy, HostListener } from '@angular/core';
import { ActivatedRoute, ParamMap } from '@angular/router';
import { ResourceService, TextSearch } from '../resource.service';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { LanguageService } from '../services/language.service';

@Component({
  selector: 'app-search',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.css']
})
export class SearchComponent implements OnInit, OnDestroy {
  public keywords:string;
  public list:any;
  public u$:Subject<void> = new Subject<void>();
  constructor(private route:ActivatedRoute,private rs:ResourceService,public ls:LanguageService) {
      this.route.paramMap.pipe(takeUntil(this.u$)).subscribe((paramMap:ParamMap)=>{
        this.keywords = paramMap.get('q').split("-").join(" ");
        this.search();
      });
      this.ls.getLanguageLabels();
   }
  public search() {
    let r = new TextSearch();
    r.setText(this.keywords);
    this.rs.shopClient.searchArticles(r,{},(e,r)=>{
      if(!e){
        this.list=[];
        r.getListList().forEach((i:any)=>{
          this.list.push({
            id:i.getId(),
            title:i.getTitle(),
            description:i.getDescription(),
            price:i.getPrice(),
            published:i.getPublished(),
            promoted:i.getPromoted(),
            images:this.rs.getLinksOfArticle(i.getImagesList()),
            videos:this.rs.getLinksOfArticle(i.getVideosList())
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
  }
  ngOnDestroy() {
    this.u$.next();
    this.u$.unsubscribe();
  }
}
