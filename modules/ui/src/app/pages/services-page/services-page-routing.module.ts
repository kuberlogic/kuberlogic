import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ServicesPageComponent } from '@app/pages/services-page/services-page.component';

const routes: Routes = [
    {
        path: '',
        component: ServicesPageComponent,
        children: [
            {
                path: '',
                loadChildren: () => import('./pages/services-list/services-list.module')
                    .then((mod) => mod.ServicesListModule),
            },
            {
                path: 'create',
                loadChildren: () => import('./pages/create-service/create-service.module')
                    .then((mod) => mod.CreateServiceModule),
            },
            {
                path: ':id',
                loadChildren: () => import('./pages/view-service/view-service.module')
                    .then((mod) => mod.ViewServiceModule),
            }
        ]
    }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule],
})
export class ServicesPageRoutingModule { }
