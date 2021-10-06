import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatTabsModule } from '@angular/material/tabs';
import { ViewServiceHeaderModule } from '@pages/services-page/pages/view-service/components/view-service-header/view-service-header.module';
import { ViewServiceRoutingModule } from '@pages/services-page/pages/view-service/view-service-routing.module';
import { ViewServiceComponent } from './view-service.component';

@NgModule({
    declarations: [ViewServiceComponent],
    imports: [
        CommonModule,
        ViewServiceRoutingModule,
        ViewServiceHeaderModule,
        MatTabsModule,
    ]
})
export class ViewServiceModule { }
