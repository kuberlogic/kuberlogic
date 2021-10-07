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

import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChangeDetectionStrategy, NO_ERRORS_SCHEMA } from '@angular/core';
import { By } from '@angular/platform-browser';
import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { ServiceConnectionTableComponent } from './service-connection-table.component';
import { ServicesPageService } from '@services/services-page.service';
import { MockServicesPageService } from '@testing/mock-services-page-service';
import { MessagesService } from '@services/messages.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MatDialog } from '@angular/material/dialog';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { FormBuilder } from '@angular/forms';
import { ServiceUsersService } from '@services/service-users.service';
import { MockServiceUsersService } from '@testing/mock-service-users-service';

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

describe('ServiceConnectionTableComponent', () => {
    let component: ServiceConnectionTableComponent;
    let fixture: ComponentFixture<ServiceConnectionTableComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceConnectionTableComponent],
            providers: [
                { provide: ServicesPageService, useClass: MockServicesPageService },
                { provide: ServiceUsersService, useClass: MockServiceUsersService },
                { provide: MessagesService, useClass: MockMessageService },
                { provide: MatDialog, useClass: MockMatDialog },
                FormBuilder,
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).overrideComponent(ServiceConnectionTableComponent, {
            set: {  changeDetection: ChangeDetectionStrategy.Default  }
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceConnectionTableComponent);
        component = fixture.componentInstance;
        component.serviceModel = serviceModel;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show external connection host when "showInnerConnection" is "false"', () => {
        component.showInnerConnection = false;
        fixture.detectChanges();

        const host = fixture.debugElement.query(By.css('.service-connection-table__host'));
        expect(host.nativeElement.textContent).toContain('external-pg-master-cloudlinux-a892.domain.com');
        expect(component).toBeTruthy();
    });

    it('should show inner connection host when "showInnerConnection" is "true"', () => {
        component.showInnerConnection = true;
        fixture.detectChanges();

        const host = fixture.debugElement.query(By.css('.service-connection-table__host'));
        expect(host.nativeElement.textContent).toContain('pg-master-cloudlinux-a892.domain.com');
        expect(component).toBeTruthy();
    });
});
