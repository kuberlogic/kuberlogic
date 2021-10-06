import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MatDialog } from '@angular/material/dialog';
import { RestoreServiceBackupsDialogComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/restore-service-backups-dialog/restore-service-backups-dialog.component';
import { MessagesService } from '@services/messages.service';
import { ServiceBackupsService } from '@services/service-backups.service';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServiceBackupsService } from '@testing/mock-service-backups-service';
import { ServiceBackupsComponent } from './service-backups.component';

describe('ServiceBackupsComponent', () => {
    let component: ServiceBackupsComponent;
    let fixture: ComponentFixture<ServiceBackupsComponent>;
    let backupsService: MockServiceBackupsService;
    let dialog: MockMatDialog;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceBackupsComponent],
            schemas: [NO_ERRORS_SCHEMA],
            imports: [HttpClientTestingModule],
            providers: [
                { provide: ServiceBackupsService, useClass: MockServiceBackupsService },
                { provide: MatDialog, useClass: MockMatDialog },
                { provide: MessagesService, useClass: MockMessageService },
                RestoreServiceBackupsDialogComponent,
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceBackupsComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        backupsService = TestBed.inject(ServiceBackupsService);
        // @ts-ignore
        dialog = TestBed.inject(MatDialog);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open "RestoreServiceBackupsDialogComponent" dialog on restore', () => {
        const spy = spyOn(dialog, 'open').and.callThrough();
        component.onRestore('db1');
        fixture.detectChanges();

        // @ts-ignore
        expect(spy.calls.mostRecent().args[0]).toEqual(RestoreServiceBackupsDialogComponent);
    });
});
