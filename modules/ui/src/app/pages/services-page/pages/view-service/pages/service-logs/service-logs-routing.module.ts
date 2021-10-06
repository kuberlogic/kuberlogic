import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ServiceLogsComponent } from '@pages/services-page/pages/view-service/pages/service-logs/service-logs.component';

const routes: Routes = [
    { path: '', component: ServiceLogsComponent }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule],
})
export class ServiceLogsRoutingModule { }
