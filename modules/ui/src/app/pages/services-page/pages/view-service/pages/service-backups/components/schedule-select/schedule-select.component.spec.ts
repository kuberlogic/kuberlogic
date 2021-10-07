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
import { MatDialog } from '@angular/material/dialog';
import { MatTableModule } from '@angular/material/table';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { NgxFilesizeModule } from 'ngx-filesize';
import { ScheduleSelectComponent, ScheduleType } from './schedule-select.component';

describe('ScheduleSelectComponent', () => {
    let component: ScheduleSelectComponent;
    let fixture: ComponentFixture<ScheduleSelectComponent>;
    let dialog: MockMatDialog;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                MatTableModule,
                NgxFilesizeModule,
            ],
            declarations: [ScheduleSelectComponent],
            providers: [
                { provide: MatDialog, useClass: MockMatDialog },
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ScheduleSelectComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        dialog = TestBed.inject(MatDialog);

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should emit on onChangeHour', () => {
        const spy = spyOn(component.saveNeeded, 'emit').and.callThrough();
        component.onChangeHour(4);

        expect(spy).toHaveBeenCalled();
    });

    it('should emit on onChangeWeekday', () => {
        const spy = spyOn(component.saveNeeded, 'emit').and.callThrough();
        component.onChangeWeekday(4);

        expect(spy).toHaveBeenCalled();
    });

    it('should emit on onChangeType daily', () => {
        const spy = spyOn(component.saveNeeded, 'emit').and.callThrough();
        component.onChangeType(ScheduleType.DAILY);

        expect(spy).toHaveBeenCalled();
        expect(component.selectedWeekday).toEqual('*');
    });

    it('should emit on onChangeType weekly', () => {
        const spy = spyOn(component.saveNeeded, 'emit').and.callThrough();
        component.onChangeType(ScheduleType.WEEKLY);

        expect(spy).toHaveBeenCalled();
        expect(component.selectedWeekday).toEqual(0);
    });

    it('should update selected values on writeValue with weekly cron', () => {
        // const spy = spyOn(component.saveNeeded, 'emit').and.callThrough();
        component.writeValue(`0 3 * * 5`);
        fixture.detectChanges();
        expect(component.selectedHour).toEqual(3);
        expect(component.selectedWeekday).toEqual(5);
        expect(component.selectedType).toEqual(ScheduleType.WEEKLY);
    });

    it('should update selected values on writeValue with daily cron', () => {
        // const spy = spyOn(component.saveNeeded, 'emit').and.callThrough();
        component.writeValue(`0 4 * * *`);
        fixture.detectChanges();
        expect(component.selectedHour).toEqual(4);
        expect(component.selectedWeekday).toEqual('*');
        expect(component.selectedType).toEqual(ScheduleType.DAILY);
    });

    it('should set default values on empty writeValue', () => {
        // const spy = spyOn(component.saveNeeded, 'emit').and.callThrough();
        component.writeValue('');
        fixture.detectChanges();
        expect(component.selectedHour).toEqual(2);
        expect(component.selectedWeekday).toEqual(0);
        expect(component.selectedType).toEqual(ScheduleType.WEEKLY);
    });
});
