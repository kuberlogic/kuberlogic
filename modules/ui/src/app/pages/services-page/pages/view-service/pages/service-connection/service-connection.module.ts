import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ServiceConnectionTableModule } from '@pages/services-page/pages/view-service/pages/service-connection/components/service-connection-table/service-connection-table.module';
import { ServiceConnectionRoutingModule } from '@pages/services-page/pages/view-service/pages/service-connection/service-connection-routing.module';
import { ServiceConnectionComponent } from './service-connection.component';

@NgModule({
    declarations: [ServiceConnectionComponent],
    imports: [
        CommonModule,
        ServiceConnectionRoutingModule,
        MatIconModule,
        MatSlideToggleModule,
        MatTooltipModule,
        ServiceConnectionTableModule,
        FormsModule
    ]
})
export class ServiceConnectionModule { }
