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

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ComponentFixture, fakeAsync, TestBed, tick } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { ServiceModel, ServiceModelType } from '@models/service.model';
import { ServicesListComponent } from '@pages/services-page/pages/services-list/services-list.component';
import { MessagesService } from '@services/messages.service';
import { ServicesPageService } from '@services/services-page.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServicesPageService } from '@testing/mock-services-page-service';

import { CreateServiceComponent } from './create-service.component';

const serviceModel: Partial<ServiceModel> = {
    type: ServiceModelType.POSTGRES,
    name: 'postgres',
    ns: 'default',
    version: '13',
    limits: {
        cpu: '1',
        memory: '1G',
        volumeSize: '10G'
    }
};

describe('CreateServiceComponent', () => {
    let component: CreateServiceComponent;
    let fixture: ComponentFixture<CreateServiceComponent>;
    let messagesService: MockMessageService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [RouterTestingModule.withRoutes(
                [
                    {path: 'services', component: ServicesListComponent},
                ]
            )],
            declarations: [CreateServiceComponent],
            providers: [
                { provide: MessagesService, useClass: MockMessageService },
                { provide: ServicesPageService, useClass: MockServicesPageService },
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateServiceComponent);
        component = fixture.componentInstance;
        messagesService = TestBed.inject(MessagesService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show success message if create is successful', fakeAsync(() => {
        const spy = spyOn(messagesService, 'success');
        component.createService(serviceModel);
        tick(10000);

        expect(spy).toHaveBeenCalled();
    }));
});
