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

import { ChangeDetectionStrategy, Component, forwardRef, Input, OnInit } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { RadioGroupModel } from '@ui/radio-group/radio-group.model';

const TYPE_CONTROL_ACCESSOR = {
    provide: NG_VALUE_ACCESSOR,
    useExisting: forwardRef(() => RadioGroupComponent),
    multi: true
};

@Component({
    selector: 'kl-radio-group',
    templateUrl: './radio-group.component.html',
    styleUrls: ['./radio-group.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    providers: [TYPE_CONTROL_ACCESSOR]
})
export class RadioGroupComponent implements OnInit, ControlValueAccessor {
    @Input() selectors: RadioGroupModel[] = [];
    value: string | number | undefined = undefined;

    constructor() { }

    ngOnInit(): void {
    }

    registerOnChange(fn: any): void {
        this.onModelChange = fn;
    }

    registerOnTouched(fn: any): void {
        this.onTouch = fn;
    }

    writeValue(obj: any): void {
        this.value = obj;
    }

    selectType(value: string | number): void {
        this.writeValue(value);
        this.onModelChange(value);
        this.onTouch();
    }

    trackByValue(_index: number, item: RadioGroupModel): string | number {
        return item.value;
    }

    private onTouch = () => {};
    private onModelChange = (_value: any) => {};

}
