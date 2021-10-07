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
import { ComponentFixture, TestBed } from '@angular/core/testing';

import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MessagesService } from '@services/messages.service';
import { ServiceLogsService } from '@services/service-logs.service';
import { MockMessageService } from '@testing/mock-messages-service';
import { MockServiceLogsService } from '@testing/mock-service-logs-service';
import { ServiceLogsComponent } from './service-logs.component';

const typesMock: {[key: string]: string} = {
    '': '',
    postgresql: 'PostgreSQL',
    mysql: 'MySQL',
    other: 'other',
};

describe('ServiceLogsComponent', () => {
    let component: ServiceLogsComponent;
    let fixture: ComponentFixture<ServiceLogsComponent>;
    let logsService: MockServiceLogsService;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ServiceLogsComponent],
            schemas: [NO_ERRORS_SCHEMA],
            imports: [HttpClientTestingModule],
            providers: [
                { provide: ServiceLogsService, useClass: MockServiceLogsService },
                { provide: MessagesService, useClass: MockMessageService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceLogsComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        logsService = TestBed.inject(ServiceLogsService);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    Object.keys(typesMock).map((type) => {
        it(`should render type for string "${typesMock[type]}"`, () => {
            expect(component.renderType(type)).toEqual(typesMock[type]);
        });
    });

    it(`should render type for undefined`, () => {
        expect(component.renderType(undefined)).toEqual(undefined);
    });

    it('should get new logs on form submit', () => {
        const spy = spyOn(logsService, 'get').and.callThrough();
        component.onFormSubmit('some_username');
        fixture.detectChanges();

        expect(spy).toHaveBeenCalled();
    });
});
