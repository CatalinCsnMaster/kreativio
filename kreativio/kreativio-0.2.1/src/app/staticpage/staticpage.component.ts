import { Component, OnInit } from '@angular/core';
import { ResourceService } from '../resource.service';
import { CallService } from 'schema-admin';
import { takeUntil } from 'rxjs/operators';
import { Subject } from 'rxjs';
import { Router, ActivatedRoute } from '@angular/router';
import { DomSanitizer } from '@angular/platform-browser';

@Component({
  selector: 'app-staticpage',
  templateUrl: './staticpage.component.html',
  styleUrls: ['./staticpage.component.css']
})
export class StaticpageComponent implements OnInit {
  public view = {title:'',text:null};
  private u$:Subject<void> = new Subject<void>(); 
  constructor(private rs:ResourceService,private call:CallService,private active:ActivatedRoute,sanitizer:DomSanitizer) {
    this.rs.getStaticSchema("https://kreativio.ro:15001",call,"kreativio-ro");
    this.call.schema.pipe(takeUntil(this.u$)).subscribe((data)=>{
      this.active.queryParamMap.pipe(takeUntil(this.u$)).subscribe(map=>{
        let key = map.get('view').split("-").join(" ");
        data.read.continut_static.value.forEach((e:any)=>{
          if(e.title==key){
            this.view.title=e.title;
            this.view.text=sanitizer.bypassSecurityTrustHtml(e.text);
          }
        });
      });
    });
  }

  ngOnInit() {
  }

}
