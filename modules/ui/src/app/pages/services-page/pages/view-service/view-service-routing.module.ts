import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ViewServiceComponent } from '@pages/services-page/pages/view-service/view-service.component';

const routes: Routes = [
    {
        path: '', component: ViewServiceComponent,
        children: [
            {
                path: '',
                redirectTo: 'connection',
                pathMatch: 'full',
            },
            {
                path: 'connection',
                loadChildren: () => import('./pages/service-connection/service-connection.module')
                    .then((mod) => mod.ServiceConnectionModule),
            },
            {
                path: 'settings',
                loadChildren: () => import('./pages/service-settings/service-settings.module')
                    .then((mod) => mod.ServiceSettingsModule),
            },
            {
                path: 'logs',
                loadChildren: () => import('./pages/service-logs/service-logs.module')
                    .then((mod) => mod.ServiceLogsModule),
            },
            {
                path: 'backups',
                loadChildren: () => import('./pages/service-backups/service-backups.module')
                    .then((mod) => mod.ServiceBackupsModule),
            },
        ]
    }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule],
})
export class ViewServiceRoutingModule { }
