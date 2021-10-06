import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { AuthGuard } from '@pages/login/auth.guard';
import { AuthService } from '@services/auth.service';

const routes: Routes = [
    {
        path: '',
        redirectTo: 'services',
        pathMatch: 'full',
    },
    {
        path: 'services',
        canActivate: [AuthGuard],
        loadChildren: () => import('./pages/services-page/services-page.module')
            .then((mod) => mod.ServicesPageModule),
    },
    {
        path: 'login',
        canActivate: [AuthService],
        loadChildren: () => import('./pages/login/login.module')
            .then((mod) => mod.LoginModule),
    },
];

@NgModule({
    imports: [RouterModule.forRoot(routes, {useHash: true})],
    exports: [RouterModule]
})
export class AppRoutingModule { }
