import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChangeDetectionStrategy, NO_ERRORS_SCHEMA } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { BackupFormComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-form/backup-form.component';
import { BackupConfigService } from '@services/backup-config.service';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { MockBackupConfigService } from '@testing/mock-backup-config-service';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServicesPageService } from '@testing/mock-services-page-service';

describe('BackupFormComponent', () => {
    let component: BackupFormComponent;
    let fixture: ComponentFixture<BackupFormComponent>;
    let servicesPageService: MockServicesPageService;
    let backupConfigService: MockBackupConfigService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [BackupFormComponent],
            providers: [
                { provide: ServicesPageService, useClass: MockServicesPageService },
                { provide: BackupConfigService, useClass: MockBackupConfigService },
                { provide: MessagesService, useClass: MockMessageService },
                { provide: MatDialog, useClass: MockMatDialog },
                FormBuilder,
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).overrideComponent(BackupFormComponent, {
            set: {  changeDetection: ChangeDetectionStrategy.Default  }
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(BackupFormComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        servicesPageService = TestBed.inject(ServicesPageService);
        // @ts-ignore
        backupConfigService = TestBed.inject(BackupConfigService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should save service on save', () => {
        const spy = spyOn(backupConfigService, 'editBackupConfig').and.callThrough();
        component.save();

        expect(spy).toHaveBeenCalled();
    });
});
