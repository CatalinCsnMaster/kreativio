import { Component, OnInit, Inject } from '@angular/core';
import { ResourceService, CategoryListConditions, Category, CategoryList } from 'src/app/resource.service';
import { moveItemInArray, CdkDragDrop } from '@angular/cdk/drag-drop';
import { LanguageService } from 'src/app/services/language.service';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialog } from '@angular/material';

@Component({
  selector: 'app-categories',
  templateUrl: './categories.component.html',
  styleUrls: ['./categories.component.css']
})
export class CategoriesComponent implements OnInit {
  public list = [];
  public listConditions;
  fake = [
    {id:1,label:'category 1'},
    {id:2,label:'category 2'},
    {id:3,label:'category 3'},
    {id:4,label:'category 4'},
    {id:5,label:'category 5'},
    {id:6,label:'category 6'},
    {id:7,label:'category 7'}
  ];
  public language;
  private language_all;
  private u$:Subject<void> = new Subject<void>();
  constructor(private rs:ResourceService,public ls:LanguageService,public dialog:MatDialog) {
    ls.getLanguageLabels();
    this.ls.language.pipe(takeUntil(this.u$)).subscribe((data:any)=>{
      this.language_all = data;
      this.language = data.ro;
    });
    let conditions = new CategoryListConditions();
    this.rs.listCategories(conditions,(r)=>{this.asObject(r.getListList())});
    //this.asObject(this.listFake());
   }
   removeField(label:string) {
    let cp = [];
    this.list.forEach(item=>{
      if(item.label!=label) cp.push(item);
    });
    this.list = cp;
   }
   newCategory(name) {
      this.list.unshift({id:0,label:name.value});
   }
   wgrpcSerialize(listAsObjectArray) {
     let list = new CategoryList();
     let arr = [];
     listAsObjectArray.forEach(item=>{
      let c = new Category();
      c.setId(item.id);
      c.setLabel(item.label);
      arr.push(c);
     });
     list.setListList(arr);
     return list;
   }
   asObject(data):any {
     this.list = [];
    data.forEach(item=>{
      this.list.push({id:item.getId(),label:item.getLabel()});
    });
   }

   listFake() {
    let l = new CategoryList();
    let list = []
    this.fake.forEach(item=>{
      let c = new Category();
      c.setId(item.id);
      c.setLabel(item.label);
      list.push(c);
    });
    l.setListList(list);
    return l.getListList();
   }

  save() {
    let categoryList = this.wgrpcSerialize(this.list)
    categoryList.setToken(sessionStorage.getItem('t'));
    this.saveCategories(categoryList);
  }

  saveCategories(categoryList) {
    this.rs.shopClient.saveCategories(categoryList,{},(e,r)=>{
      if(!e){
        this.asObject(r.getListList());
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

  openDialog(title,i): void {
    const dialogRef = this.dialog.open(CatEdit, {
      width: '350px',
      data: {title:title},
      panelClass:'light-mode'
    });

    dialogRef.afterClosed().subscribe(result => {
      if(result) this.list[i].label = result;
    });
  }

  dropped(event){
    moveItemInArray(event.container.data,event.previousIndex,event.currentIndex)
  }

  ngOnInit() {
  }

}
@Component({
  selector: 'app-catedit',
  templateUrl: 'catedit.component.html',
})
export class CatEdit {
title:string;
  constructor(
    public dialogRef: MatDialogRef<CatEdit>,
    @Inject(MAT_DIALOG_DATA) public data:any) {
      this.title = data.title;
    }

  close(): void {
    this.dialogRef.close(this.title);
  }

}