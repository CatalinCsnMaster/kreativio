import { Component, OnInit, Input } from '@angular/core';

export interface sidebar_static_link{
    label:string;
    link:string;
    icon?:string;
}

@Component({
  selector: 'app-staticlinks',
  templateUrl: './staticlinks.component.html',
  styleUrls: ['./staticlinks.component.css']
})
export class StaticlinksComponent implements OnInit {
  constructor() {}
  public static_links:sidebar_static_link[] = [];
  ngOnInit() {}
}
