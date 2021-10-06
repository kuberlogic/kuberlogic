import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { ServicesPageRoutingModule } from '@pages/services-page/services-page-routing.module';
import { ServicesPageComponent } from './services-page.component';

@NgModule({
    declarations: [ServicesPageComponent],
    imports: [
        CommonModule,
        RouterModule,
        ServicesPageRoutingModule
    ]
})
export class ServicesPageModule { }
