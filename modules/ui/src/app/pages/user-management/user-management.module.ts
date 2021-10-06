import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { UserManagementRoutingModule } from '@app/pages/user-management/user-management-routing.module';
import { UserManagementComponent } from './user-management.component';

@NgModule({
    declarations: [UserManagementComponent],
    imports: [
        CommonModule,
        UserManagementRoutingModule,
    ]
})
export class UserManagementModule { }
