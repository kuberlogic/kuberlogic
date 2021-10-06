import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ServicesListHeaderModule } from '@pages/services-page/pages/services-list/components/services-list-header/services-list-header.module';
import { ServicesListTableModule } from '@pages/services-page/pages/services-list/components/services-list-table/services-list-table.module';
import { ServicesListRoutingModule } from '@pages/services-page/pages/services-list/services-list-routing.module';
import { ServicesListComponent } from './services-list.component';

@NgModule({
    declarations: [ServicesListComponent],
    imports: [
        CommonModule,
        ServicesListRoutingModule,
        ServicesListTableModule,
        ServicesListHeaderModule,
    ]
})
export class ServicesListModule { }
