import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { Router } from '@angular/router';
import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { MockConfirmDialogService } from '@testing/mock-confirm-dialog-service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockPipeFactory } from '@testing/mock-pipe-factory';
import { MockRouter } from '@testing/mock-router';
import { MockServicesPageService } from '@testing/mock-services-page-service';
import { ConfirmDialogService } from '@ui/confirm-dialog/confirm-dialog.service';
import { ViewServiceHeaderComponent } from './view-service-header.component';

const serviceModel: ServiceModel = {
    type: ServiceModelType.POSTGRES,
    name: 'PostgreSql prod',
    status: ServiceModelStatus.FAILED,
    masters: 2,
    replicas: 5,
    created_at: '2021-02-09T10:56:12.115Z',
    internalConnection: {
        master: {
            host: 'pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
        replica: {
            host: 'pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
    },
    externalConnection: {
        master: {
            host: 'external-pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
        replica: {
            host: 'external-pg-master-cloudlinux-a892.domain.com',
            port: 20990,
            user: 'cloudmanaged_admin',
            password: '*********',
            ssl_mode: 'require',
            cert: '/cert_link'
        },
    }
};

describe('ViewServiceHeaderComponent', () => {
    let component: ViewServiceHeaderComponent;
    let fixture: ComponentFixture<ViewServiceHeaderComponent>;
    let messagesService: MessagesService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [
                ViewServiceHeaderComponent,
                MockPipeFactory('klServiceStatusName'),
                MockPipeFactory('klServiceVersion'),
            ],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
                { provide: Router, useClass: MockRouter },
                { provide: ServicesPageService, useClass: MockServicesPageService },
                { provide: ConfirmDialogService, useClass: MockConfirmDialogService },
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ViewServiceHeaderComponent);
        component = fixture.componentInstance;
        component.serviceModel = serviceModel;
        messagesService = TestBed.inject(MessagesService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show success message after delete', () => {
        const spy = spyOn(messagesService, 'success');
        component.deleteService();
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });
});
