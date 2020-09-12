import { Component, OnInit, Output, HostListener, ChangeDetectionStrategy, ChangeDetectorRef, OnDestroy, ViewChild, Input } from '@angular/core';
import { Subject } from 'rxjs';
import { FormControl } from '@angular/forms';
import { MatSlideToggle, MatInput } from '@angular/material';
import { Router } from '@angular/router';
import { ResourceService, SuggestionList, TextSearch } from '../resource.service';

@Component({
  selector: 'app-toolbar',
  templateUrl: './toolbar.component.html',
  styleUrls: ['./toolbar.component.css']
})
export class ToolbarComponent implements OnInit,OnDestroy {
  @ViewChild('sidebarSlideToggle',{read:MatSlideToggle,static:true}) @Output() public sidebarSlider:MatSlideToggle;
  @Output() sidebarToggle:Subject<boolean> = new Subject();
  @ViewChild('searchInput',{read:MatInput,static:true}) searchInputt:MatInput;
  public suggestion_articles = [];
  public suggestion_categories = [];
  @HostListener('document:keyup',['$event']) suggest(event:KeyboardEvent){
    if(this.searchInputt.focused&&this.searchInputt.value.length>=3) {
      let req = new TextSearch();
      req.setText(`${this.searchInputt.value.split(" ")[0]}:*`);
      this.rs.shopClient.suggest(req,{},(e,r)=>{
        if(!e) {
          this.suggestion_articles = r.getArticleList();
          this.suggestion_categories = r.getCategoryList();
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
  }
  suggestion_box_hidden:boolean;

  constructor(public cdRef:ChangeDetectorRef,private router:Router,private rs:ResourceService) { }
  ngOnInit() {}
  search_input:FormControl = new FormControl();
  
  public search(event){
    if(event.keyCode==13){
      this.search_();
    }
  }
  public search_(){
    let keywords;
    if(this.search_input.value != null)  {
    keywords = this.search_input.value.split(' ').join('-');
    this.router.navigate(['refresh']).then(()=>this.router.navigate(['/search',{q:keywords}]));
    } else {
      // ...
    }
  }
  //@ViewChild('bar',{read:MatToolbar,static:true}) toolbar:MatToolbar;
   @Input() public isTop:boolean;
    ngAfterContentInit(): void {

    }
    getSuggestionBoxStatus() {
      return this.suggestion_box_hidden;
    }
    ngAfterViewChecked(): void {
      //Called after every check of the component's view. Applies to components only.
      //Add 'implements AfterViewChecked' to the class.
      if(this.searchInputt.focused) {
        this.suggestion_box_hidden = false;
      } else {
        setTimeout(()=>{this.suggestion_box_hidden = true;},1000);
      }
      this.cdRef.detectChanges()
    }
    ngOnDestroy(): void {
      this.cdRef.detach();
    }
}
