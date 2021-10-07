/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
