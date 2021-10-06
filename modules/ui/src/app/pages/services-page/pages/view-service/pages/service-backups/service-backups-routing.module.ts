import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ServiceBackupsComponent } from '@pages/services-page/pages/view-service/pages/service-backups/service-backups.component';

const routes: Routes = [
    { path: '', component: ServiceBackupsComponent }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule],
})
export class ServiceBackupsRoutingModule { }
