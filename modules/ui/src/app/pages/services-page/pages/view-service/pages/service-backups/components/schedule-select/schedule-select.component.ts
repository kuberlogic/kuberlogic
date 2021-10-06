/**
 * Copyright © Cloud Linux GmbH & Cloud Linux Software, Inc 2010-2020 All Rights Reserved
 *
 * Licensed under CLOUD LINUX LICENSE AGREEMENT
 * http://cloudlinux.com/docs/LICENSE.TXT
 */
import { Component, EventEmitter, forwardRef, Output } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { parseExpression } from 'cron-parser';

export enum ScheduleType {
    DAILY = 'Daily',
    WEEKLY = 'Weekly',
}

@Component({
    selector: 'kl-schedule-select',
    providers: [
        {
            provide: NG_VALUE_ACCESSOR,
            useExisting: forwardRef(() => ScheduleSelectComponent),
            multi: true,
        },
    ],
    styleUrls: ['./schedule-select.component.scss'],
    templateUrl: './schedule-select.component.html',
})
export class ScheduleSelectComponent implements ControlValueAccessor {
    ScheduleType = ScheduleType;
    @Output() saveNeeded = new EventEmitter<void>();
    selectedType: ScheduleType | undefined;
    selectedHour: number | '*' | undefined;
    selectedWeekday: number | '*' | undefined;
    next: any;

    readonly hours = [...Array(24).keys()];
    readonly weekdays = [
        'Sunday',
        'Monday',
        'Tuesday',
        'Wednesday',
        'Thursday',
        'Friday',
        'Saturday',
    ];
    readonly types = [ScheduleType.DAILY, ScheduleType.WEEKLY];

    save(): void {
        this.onChange(this.getCronString());
        this.saveNeeded.emit();
    }
    onChangeHour(value: any): void {
        this.selectedHour = value;
        this.save();
    }
    onChangeWeekday(value: any): void {
        this.selectedWeekday = value;
        this.save();
    }
    onChangeType(value: ScheduleType): void {
        if (value === ScheduleType.DAILY) {
            this.selectedWeekday = '*';
        }
        if (value === ScheduleType.WEEKLY) {
            this.selectedWeekday = 0;
        }
        this.save();
    }
    getCronString(): string {
        return `0 ${this.selectedHour} * * ${this.selectedWeekday}`;
    }
    getHourLabel(h: number): string {
        return `${h < 10 ? '0' : ''}${h}:00`;
    }

    onChange = (v: any) => {};
    writeValue(value: string): void {
        if (!value) {
            this.selectedHour = 2;
            this.selectedWeekday = 0;
            this.selectedType = ScheduleType.WEEKLY;
        } else {
            this.next = parseExpression(value).next();
            const regexp = /[\d*|\*] (\d+|\*) [\d*|\*] [\d*|\*] (\d+|\*)/;
            const match = value.match(regexp);
            if (match) {
                this.selectedHour = match[1] === '*' ? '*' : Number(match[1]);
                this.selectedWeekday = match[2] === '*' ? '*' : Number(match[2]);
                this.selectedType = this.selectedWeekday === '*'
                    ? ScheduleType.DAILY
                    : ScheduleType.WEEKLY;
            }
        }
    }
    registerOnTouched(): void {}
    registerOnChange(fn: any): void {
        this.onChange = fn;
    }
}
