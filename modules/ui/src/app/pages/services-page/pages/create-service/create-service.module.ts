import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { CreateServiceFormModule } from '@pages/services-page/pages/create-service/components/create-service-form/create-service-form.module';
import { CreateServiceHeaderModule } from '@pages/services-page/pages/create-service/components/create-service-header/create-service-header.module';
import { CreateServiceRoutingModule } from '@pages/services-page/pages/create-service/create-service-routing.module';
import { CreateServiceComponent } from './create-service.component';

@NgModule({
    declarations: [CreateServiceComponent],
    imports: [
        CommonModule,
        CreateServiceRoutingModule,
        CreateServiceHeaderModule,
        CreateServiceFormModule,
    ]
})
export class CreateServiceModule { }
