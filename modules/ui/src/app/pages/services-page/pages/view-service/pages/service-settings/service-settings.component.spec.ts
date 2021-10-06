import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChangeDetectionStrategy, NO_ERRORS_SCHEMA } from '@angular/core';
import { ServiceBackupConfigModel } from '@models/service-backup-config.model';
import { ServiceModel } from '@models/service.model';
import { ServiceSettingsFormResult } from '@pages/services-page/pages/view-service/pages/service-settings/components/service-settings-form/service-settings-form.component';
import { BackupConfigService } from '@services/backup-config.service';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { MockBackupConfigService } from '@testing/mock-backup-config-service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServicesPageService } from '@testing/mock-services-page-service';
import { ServiceSettingsComponent } from './service-settings.component';

const result: ServiceSettingsFormResult = {
    service : {} as ServiceModel,
};

describe('ServiceSettingsComponent', () => {
    let component: ServiceSettingsComponent;
    let fixture: ComponentFixture<ServiceSettingsComponent>;
    let servicesPageService: MockServicesPageService;
    let backupConfigService: MockBackupConfigService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceSettingsComponent],
            providers: [
                { provide: ServicesPageService, useClass: MockServicesPageService },
                { provide: BackupConfigService, useClass: MockBackupConfigService },
                { provide: MessagesService, useClass: MockMessageService },
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).overrideComponent(ServiceSettingsComponent, {
            set: {  changeDetection: ChangeDetectionStrategy.Default  }
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceSettingsComponent);
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

    it('should save service onSave', () => {
        const spy = spyOn(servicesPageService, 'editService').and.callThrough();
        component.onSave(result);

        expect(spy).toHaveBeenCalled();
    });
});
