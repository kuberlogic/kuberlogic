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

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MockMatDialogRef } from '@testing/mock-mat-dialog-ref';
import { ConfirmDialogComponent } from './confirm-dialog.component';

const MockDialogData = {
    title: 'title',
    message: 'message',
    buttonConfirmText: 'Yes',
    buttonCancelText: 'No',
};

describe('ConfirmDialogComponent', () => {
    let component: ConfirmDialogComponent;
    let fixture: ComponentFixture<ConfirmDialogComponent>;
    let dialogRef: MockMatDialogRef;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ConfirmDialogComponent],
            providers: [
                { provide: MatDialogRef, useClass: MockMatDialogRef },
                { provide: MAT_DIALOG_DATA, useValue: MockDialogData }
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfirmDialogComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        dialogRef = TestBed.inject(MatDialogRef);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should close dialog with "true" value onConfirm', () => {
        const spy = spyOn(dialogRef, 'close');
        component.onConfirm();

        expect(spy).toHaveBeenCalledWith(true);
    });

    it('should close dialog with "false" value onDismiss', () => {
        const spy = spyOn(dialogRef, 'close');
        component.onDismiss();

        expect(spy).toHaveBeenCalledWith(false);
    });
});
