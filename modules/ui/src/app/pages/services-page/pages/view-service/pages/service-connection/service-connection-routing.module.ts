import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ServiceConnectionComponent } from '@pages/services-page/pages/view-service/pages/service-connection/service-connection.component';

const routes: Routes = [
    { path: '', component: ServiceConnectionComponent }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule],
})
export class ServiceConnectionRoutingModule { }
