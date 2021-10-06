import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { ServiceLogsFormModule } from '@pages/services-page/pages/view-service/pages/service-logs/components/service-logs-form/service-logs-form.module';
import { ServiceLogsRoutingModule } from '@pages/services-page/pages/view-service/pages/service-logs/service-logs-routing.module';
import { ServiceLogsComponent } from './service-logs.component';

@NgModule({
    declarations: [ServiceLogsComponent],
    imports: [
        CommonModule,
        MatIconModule,
        ServiceLogsRoutingModule,
        ServiceLogsFormModule,
    ]
})
export class ServiceLogsModule { }
