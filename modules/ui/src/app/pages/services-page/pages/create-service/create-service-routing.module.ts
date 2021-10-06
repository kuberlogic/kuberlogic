import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { CreateServiceComponent } from '@pages/services-page/pages/create-service/create-service.component';

const routes: Routes = [
    { path: '', component: CreateServiceComponent }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule],
})
export class CreateServiceRoutingModule { }
