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
import { ServiceRestoresTableComponent } from './service-restores-table.component';

describe('ServiceRestoresTableComponent', () => {
    let component: ServiceRestoresTableComponent;
    let fixture: ComponentFixture<ServiceRestoresTableComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [],
            declarations: [ServiceRestoresTableComponent],
            providers: [],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ServiceRestoresTableComponent);
        component = fixture.componentInstance;
        component.restores = [
            {
                file: 's3://test/postgresql/kuberlogic-kl-pg/logical_backups/1622729871.sql.gz',
                database: 'db1',
                time: '2021-06-04T08:00:14.000Z',
                status: 'Failed',
            },
        ];
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
