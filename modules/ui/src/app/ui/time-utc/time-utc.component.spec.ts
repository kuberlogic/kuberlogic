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

import { TimeagoModule } from 'ngx-timeago';
import { TimeUtcComponent } from './time-utc.component';

describe('TimeUtcComponent', () => {
    let component: TimeUtcComponent;
    let fixture: ComponentFixture<TimeUtcComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                TimeagoModule.forRoot(),
            ],
            declarations: [
                TimeUtcComponent,
            ],
        })
            .compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TimeUtcComponent);
        component = fixture.componentInstance;
        component.timestamp = '2021-06-04T00:00:13.794Z';
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
