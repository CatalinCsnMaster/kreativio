import { Component, Inject, AfterContentChecked, AfterContentInit } from "@angular/core";
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material';
import { CartService } from '../cart.service';
import { ResourceService } from '../resource.service';
import { LanguageService } from '../services/language.service';

@Component({
    selector:'cart-dialog',
    styles:[],
    template:`
        <div>
        <h3 mat-dialog-title> 
             <a *ngIf="data.list.length!=0" style="color:grey;"></a>
             <p *ngIf="data.list.length==0">Cos de cumparaturi gol.</p>
        </h3>
        <div *ngIf="data.list.length!=0" class="text-center rounded shadow" mat-dialog-content>
            
            <table mat-table [dataSource]="data.list" class="mat-elevation-z8">

            <ng-container matColumnDef="title">
                <th mat-header-cell *matHeaderCellDef>Articol&nbsp;&nbsp;&nbsp;</th>
                <td mat-cell *matCellDef="let element">{{element.title}}</td>
            </ng-container>

            <ng-container matColumnDef="quantity">
            <th mat-header-cell *matHeaderCellDef>Cantitate&nbsp;&nbsp;&nbsp;</th>
            <td mat-cell *matCellDef="let element" class="text-center">{{element.quantity}}</td>
            </ng-container>

            <ng-container matColumnDef="price_t">
            <th mat-header-cell *matHeaderCellDef>Pret</th>
            <td mat-cell *matCellDef="let element">{{element.price_t.toFixed(2)}} Lei</td>
            </ng-container>

                <tr mat-header-row *matHeaderRowDef="column_order"></tr>
                <tr mat-row *matRowDef="let row; columns: column_order;"></tr>
            </table>
            
        </div>    
        <div class="custom-theme" mat-dialog-actions>
            <button mat-raised-button class="custom-theme" color="warn" (click)="emptyCart()"><i class="material-icons">remove_shopping_cart</i>
            {{label_btn_empty}}</button>
            <a mat-raised-button color="" style="text-decoration: none;color:white;" class="custom-theme" routerLink="/cart"><i class="material-icons">shopping_cart</i>
            {{label_btn_cart}}</a>
            <button mat-raised-button color="accent" [disabled]="data.list.length==0" class="custom-theme" [mat-dialog-close]="data" (click)="checkoutCart()" cdkFocusInitial><i class="material-icons">check_circle</i>
            {{label_btn_checkout}}</button>
        </div>
        </div>
    `
})
export class CartDialog implements AfterContentInit {
    column_order = ['title','quantity','price_t'];

    label_btn_empty:string;
    label_btn_cart:string;
    label_btn_checkout:string;
    label_dialog_title:string;
    constructor
    (public dialogRef:MatDialogRef<CartDialog>,
        @Inject(MAT_DIALOG_DATA) public data,
        private cart:CartService,
        public rs:ResourceService,
        public ls:LanguageService)
    {}
    public emptyCart():void{
        this.cart.save_cart({data:[]});
        this.data.list = [];
        this.rs.notify(this.data.language.snb_emptied,3000);
    }
    public checkoutCart():void{
        this.dialogRef.close(true);
    }
    ngAfterContentInit():void{
        this.label_dialog_title = this.data.language.label_cart_dialog_title;
        this.label_btn_empty = this.data.language.label_cart_dialog_empty;
        this.label_btn_cart = this.data.language.label_cart_dialog_cart;
        this.label_btn_checkout = this.data.language.label_cart_dialog_checkout;
    }
}