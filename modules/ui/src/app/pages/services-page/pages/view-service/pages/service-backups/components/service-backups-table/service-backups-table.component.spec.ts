import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialog } from '@angular/material/dialog';
import { MatTableModule } from '@angular/material/table';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { NgxFilesizeModule } from 'ngx-filesize';
import { ServiceBackupsTableComponent } from './service-backups-table.component';

describe('ServiceBackupsTableComponent', () => {
    let component: ServiceBackupsTableComponent;
    let fixture: ComponentFixture<ServiceBackupsTableComponent>;
    let dialog: MockMatDialog;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                MatTableModule,
                NgxFilesizeModule,
            ],
            declarations: [ServiceBackupsTableComponent],
            providers: [
                { provide: MatDialog, useClass: MockMatDialog },
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceBackupsTableComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        dialog = TestBed.inject(MatDialog);
        component.backups = [{
            file: 's3://test/postgresql/kuberlogic-kl-pg/logical_backups/1622729871.sql.gz',
            size: 14021,
            lastModified: '2021-06-03T14:17:55.510Z',
        }];
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should emit "onRestore" on restore', () => {
        const spy = spyOn(component.restore, 'emit');
        component.onRestore('name');

        expect(spy).toHaveBeenCalled();
    });
});
