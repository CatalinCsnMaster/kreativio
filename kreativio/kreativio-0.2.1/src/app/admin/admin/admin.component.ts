import { Component, OnInit, ViewChild, AfterContentInit, OnDestroy } from '@angular/core';
import { CallService, ImageService } from 'schema-admin';
import { MatSlideToggle, MatSidenav } from '@angular/material';
import { LanguageService } from 'src/app/services/language.service';
import { takeUntil } from 'rxjs/operators';
import { Subject } from 'rxjs';
import { Router } from '@angular/router';
import { PlatformService } from 'src/app/services/platform.service';

@Component({
  selector: 'app-admin',
  templateUrl: './admin.component.html',
  styleUrls: ['./admin.component.css'],
  host: {
    class: 'light-mode bg-light2'
  }
})
export class AdminComponent implements OnInit,AfterContentInit,OnDestroy { 
  public language_all;
  public language;
  private u$:Subject<void> = new Subject<void>();
  @ViewChild("adminSidebarToggle",{read:MatSlideToggle,static:true}) public adminSidebarToggle:MatSlideToggle;
  @ViewChild("adminDrawer",{read:MatSidenav,static:true}) public adminSidenav:MatSidenav;
  constructor(private call:CallService,private lang_service:LanguageService,private is:ImageService,private router:Router,public platform:PlatformService) { 
    is.image_client_ept = "https://kreativio.ro:15002"
    call.schema_ept_1 = 'https://kreativio.ro:15001';
    call.lbl_btn_add_field = 'Adauga camp nou';
    call.lbl_btn_remove_img = 'Sterge';
    call.lbl_btn_remove_img_tooltip = 'Sterge imaginea definitiv.';
    is.lbl_logged_in = 'Logat cu success.';
    is.lbl_removeImg = 'Element sters.';
    call.lbl_btn_save = 'Salveaza';
    call.lbl_btn_save_tooltip = 'Salveaza modificarile facute pe aceasta pagina.';
    call.lbl_btn_remove_tooltip = 'Sterge randul.';
    //this.router.navigate(['admin','kreativio-ro']); // schema-admin redirects to base admin from there this component redirects to right schema.
  }

  ngOnInit() {
  }
  ngAfterContentInit() {
    this.adminSidebarToggle.toggle();
    this.adminSidenav.toggle();
    this.call.snackbar_close = '❌';
    this.is.snackbar_close = '❌';
    this.lang_service.language.pipe(takeUntil(this.u$)).subscribe((data:any)=>{
      this.language_all = data;
      this.language = data.ro;
    });
  }
  logout(){
    localStorage.clear();
    sessionStorage.clear();
    location.replace('/');
  }
  ngOnDestroy(){
    this.u$.next();
    this.u$.unsubscribe();
  }
}
