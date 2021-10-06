import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ServiceSettingsFormModule } from '@pages/services-page/pages/view-service/pages/service-settings/components/service-settings-form/service-settings-form.module';
import { ServiceSettingsRoutingModule } from '@pages/services-page/pages/view-service/pages/service-settings/service-settings-routing.module';
import { ServiceSettingsComponent } from './service-settings.component';

@NgModule({
    declarations: [ServiceSettingsComponent],
    imports: [
        CommonModule,
        ServiceSettingsRoutingModule,
        ServiceSettingsFormModule,
    ]
})
export class ServiceSettingsModule { }
