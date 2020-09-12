import { Component, ViewChild, AfterContentInit, OnDestroy, OnInit, ElementRef, HostListener, Inject } from '@angular/core';
import { MatSidenav, MatDialog, MatDialogRef, MAT_DIALOG_DATA, MatBottomSheetRef, MAT_BOTTOM_SHEET_DATA, MatBottomSheet } from '@angular/material';
import { takeUntil, take } from 'rxjs/operators'
import { Subject } from 'rxjs';
import { ToolbarComponent } from '../toolbar/toolbar.component';
import { CartDialog } from '../main/cart.component';
import { StaticlinksComponent, sidebar_static_link } from '../staticlinks/staticlinks.component';
import { LanguageService } from '../services/language.service';
import { PlatformService } from '../services/platform.service';
import { Router, Scroll } from '@angular/router';
import { CartService } from '../cart.service';
import { CallService, ImageService } from 'schema-admin';
import { ResourceService, CategoryListConditions, Message, ShopClient } from '../resource.service';

@Component({
  selector: 'app-customer',
  templateUrl: './customer.component.html',
  styleUrls: ['./customer.component.css']
})
export class CustomerComponent implements OnInit,AfterContentInit,OnDestroy {
  public static_schema;
  @ViewChild('drawer',{read:MatSidenav,static:true}) drawer:MatSidenav;
  @ViewChild('toolbar',{read:ToolbarComponent,static:true}) toolbar:ToolbarComponent;
  @ViewChild('staticlinks',{read:StaticlinksComponent,static:true}) sidebar_links:StaticlinksComponent;
  constructor(public platform:PlatformService,
    private cartDialog:MatDialog,
    private lang_service:LanguageService,
    private router:Router,
    private cart:CartService,
    private call:CallService,
    private rs:ResourceService,
    private contactDialog:MatBottomSheet) {
      this.rs.getStaticSchema("https://kreativio.ro:15001",call,"kreativio-ro");
      this.call.schema.pipe(takeUntil(this.u$)).subscribe((data)=>{
        this.static_schema = data.read;
        let static_lnks:sidebar_static_link[] = [];
        for(let i=0;i<data.read.continut_static.value.length;i++){
        static_lnks.push({label:`${data.read.continut_static.value[i].button_label}`,link:`${data.read.continut_static.value[i].title.split(" ").join("-")}`,icon:""});
        }
        this.sidebar_links.static_links = static_lnks;
      });  
      this.getCategoryList();   
   }
   public mobile:boolean = false;
   private u$:Subject<void> = new Subject();
   public language;
   private language_all;
   public cart_badge_value:string;
   public drawer_mode = 'side';
  ngOnInit() {
    if(this.platform.mobile){
      this.drawer_mode = 'push';
    }
  }
  ifMobileDismissSide() {
    if(this.mobile) this.drawer.toggle();
  }
  public categories;
  getCategoryList() {
    let c = new CategoryListConditions();
    c.setOnlyPublishedArticles(true);
    this.rs.listCategories(c,(r)=>{
      this.categories = r.getListList();
    })
  }
  ngAfterContentInit(): void {
    this.toolbar.sidebarToggle.pipe(takeUntil(this.u$)).subscribe((data)=>{
      this.drawer.toggle();
    });
    this.lang_service.language.pipe(takeUntil(this.u$)).subscribe((data:any)=>{
      this.language_all = data;
      this.language = data.ro;
    });
    setInterval(()=>{if(this.cart.parse_cart().data.length>0){
      this.cart_badge_value = this.cart.parse_cart().data.length;
    }else{this.cart_badge_value=''}},3000);

    if(this.contentContainer.nativeElement.scrollTop==0&&!this.platform.mobile)this.toolbar.isTop = true;

    this.drawer.closedStart.pipe(takeUntil(this.u$)).subscribe(data=>{
      if(this.toolbar.sidebarSlider.checked) this.toolbar.sidebarSlider.toggle();
    });
  }
  
  switchLanguage(k:string){
    this.language = this.language_all[k];
  }

  openCartDialog(){
    this.cartDialog.open(CartDialog,{
      maxWidth:'430px',
      maxHeight:'750px',
      // position: {top:'100px',left:'100px'},
      data:{language:this.language,list:this.cart.parse_cart().data}
    }).afterClosed().pipe(takeUntil(this.u$)).subscribe(
      (cartData)=>{
        if(cartData) this.router.navigate(['checkout']);
      }
    )
  }
  sidebarOpened:boolean = false;
  @ViewChild('contentContainer',{read:ElementRef,static:true}) public contentContainer:ElementRef;
  opened:boolean = false;
  scrolled(ev:Scroll){
    if(this.contentContainer.nativeElement.scrollTop==0&&!this.platform.mobile&&this.toolbar.sidebarSlider.checked){
      if(this.toolbar.sidebarSlider.checked){
      //this.toolbar.sidebarToggle.next(false);
      //this.toolbar.sidebarSlider.toggle();
      //this.sidebarOpened = false;
      }
        this.toolbar.isTop = true;
      }
    else if(this.contentContainer.nativeElement.scrollTop>700&&!this.platform.mobile){
        if(!this.toolbar.sidebarSlider.checked){
          if(!this.opened){
          this.toolbar.sidebarToggle.next(true);
          this.toolbar.sidebarSlider.toggle();
          this.sidebarOpened = true;
          this.opened = true;
          }
        }
          this.toolbar.isTop = false;
    }else{
      this.toolbar.isTop = false;
    }
  }
  ngOnDestroy(): void {
    this.u$.next();
    this.u$.unsubscribe();
  }
  openContactDialog(){
    this.contactDialog.open(AppContact,{data:{schema:this.static_schema,language:this.language}});
  }
}
@Component(
  {selector:'app-contact',
  template:`
  <div mat-dialog-title><a class="italiano" style="font-size:1.5em;">Kreativio.ro contact</a></div>
  <div mat-dialog-content>
  <p>{{data.schema.contact.value}}</p>
  <mat-divider></mat-divider>
 
    <div *ngFor="let item of data.schema.program.value">
    <div align="left"><small>{{item.zile}}</small></div><div align="right">{{item.orar}}</div>
    <mat-divider></mat-divider>
    </div>
  </div>
  
  <div mat-dialog-title>Trimite un mesaj</div>
    <div mat-dialog-content align="center">
      <mat-form-field>
        <mat-label>Prenume</mat-label>
        <input matInput #fname type="text" >
      </mat-form-field>
      <mat-form-field>
        <mat-label>Nume</mat-label>
        <input matInput #lname type="text" >
      </mat-form-field><br>
      <mat-form-field>
        <mat-label>E-mail</mat-label>
        <input matInput #email type="text" >
      </mat-form-field>
      <mat-form-field>
        <mat-label>Telefon contact</mat-label>
        <input matInput #phone type="text" >
      </mat-form-field><br>
      <mat-form-field style="width:99%">
        <mat-label>Mesaj</mat-label>
        <textarea matInput #message style="width:99%" type="text"></textarea>
      </mat-form-field>
    </div>
    <div mat-dialog-actions>
    <button mat-raised-button (click)="Close()">Inchide</button>
    <button mat-raised-button [disabled]="fname.value==''||lname.value==''||email.value<'6'||phone.value.length<6||message.value==''" 
    (click)="Contact(fname.value,lname.value,email.value,phone.value,message.value)">Trimite mesaj</button>
    </div>
  `}
)
export class AppContact{
  private shopClient;
  constructor(private dialogRef:MatBottomSheetRef<AppContact>,@Inject(MAT_BOTTOM_SHEET_DATA) public data:any,private auth:ImageService,private rs:ResourceService){
    this.shopClient = new ShopClient(this.rs.shopServiceEndpoint);
    //console.log(data);
  }
  Close(){
    this.dialogRef.dismiss();
  }

  Contact(fname,lname,email,phone,message) {
    let rq = new Message();
    rq.setName(`${fname} ${lname}`);
    rq.setEmail(email);
    rq.setPhone(phone);
    rq.setSubject(`Contact e-mail from ${fname} ${lname}`);
    rq.setMessage(message);
    this.shopClient.sendMessage(rq,{},(e,r)=>{
      if(!e) {
        return r;
      } else {
        this.rs.errorSwitch(e);
      }
    });
  }

  Contact_old(fname,lname,email,phone,message) {
    this.auth.contactMail(
      'Contact',
      fname,
      lname,
      email,
      message,
      phone).then(data=>{
        this.rs.notify('Mesaj de contact trimis!',9000);
      },
      (error)=>{
        console.log(error);
        this.rs.notify('Incercare de a trimite esuata, va rugam incercati mai tarziu.',9000);
      });
  }
}