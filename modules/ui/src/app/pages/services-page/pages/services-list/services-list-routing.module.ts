import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ServicesListComponent } from '@app/pages/services-page/pages/services-list/services-list.component';

const routes: Routes = [
    { path: '', component: ServicesListComponent }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule],
})
export class ServicesListRoutingModule { }
