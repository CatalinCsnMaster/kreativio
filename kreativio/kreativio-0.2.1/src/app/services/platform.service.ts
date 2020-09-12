import { Injectable } from '@angular/core';
import { Platform } from '@angular/cdk/platform';

@Injectable({
  providedIn: 'root'
})
export class PlatformService {
  public mobile:boolean = false;
  constructor(private platform:Platform) {
    if(this.platform.ANDROID||this.platform.IOS){
      this.mobile=true;
    }
   }

}
