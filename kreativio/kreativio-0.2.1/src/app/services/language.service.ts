import { Injectable, OnDestroy } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { takeUntil } from 'rxjs/operators';
import { Subject } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class LanguageService implements OnDestroy {
  private u$:Subject<void> = new Subject<void>();
  public language:Subject<any>=new Subject<any>();
  constructor(private http:HttpClient) {
    this.getLanguageLabels();
   }
  getLanguageLabels(cb?,k?:string):void {
    this.http.get('/assets/language.json',{responseType:'json'}).pipe(takeUntil(this.u$)).subscribe((data:any)=>{
      this.language.next(data);
      if(cb&&k)cb(k);
    });
  }
  ngOnDestroy(): void {
    this.u$.next();
    this.u$.unsubscribe();
  }
}
