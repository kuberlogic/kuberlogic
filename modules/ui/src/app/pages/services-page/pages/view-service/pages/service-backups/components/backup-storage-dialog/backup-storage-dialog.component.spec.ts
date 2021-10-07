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
import { FormBuilder } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { BackupStorageDialogComponent } from '@pages/services-page/pages/view-service/pages/service-backups/components/backup-storage-dialog/backup-storage-dialog.component';
import { MockMatDialogRef } from '@testing/mock-mat-dialog-ref';

const MockDialogData = {
    aws_access_key_id: 'aws_access_key_id',
    aws_secret_access_key: 'aws_secret_access_key',
    bucket: 'bucket',
    endpoint: 'endpoint',
};

describe('BackupStorageDialogComponent', () => {
    let component: BackupStorageDialogComponent;
    let fixture: ComponentFixture<BackupStorageDialogComponent>;
    let dialogRef: MockMatDialogRef;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [BackupStorageDialogComponent],
            providers: [
                FormBuilder,
                { provide: MatDialogRef, useClass: MockMatDialogRef },
                { provide: MAT_DIALOG_DATA, useValue: MockDialogData }
            ],
            schemas: [NO_ERRORS_SCHEMA]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(BackupStorageDialogComponent);
        component = fixture.componentInstance;
        // @ts-ignore
        dialogRef = TestBed.inject(MatDialogRef);
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    describe('when form is invalid', () => {
        beforeEach(() => {
            component.formGroup.patchValue({aws_access_key_id: ''});
            fixture.detectChanges();
        });

        it('should not emit form value', () => {
            const spy = spyOn(dialogRef, 'close');
            component.onSave();
            fixture.detectChanges();

            expect(spy).not.toHaveBeenCalled();
        });
    });

    describe('when form is valid', () => {
        const data = {
            aws_access_key_id: 'aws_access_key_id',
            aws_secret_access_key: 'aws_secret_access_key',
            bucket: 'bucket',
            endpoint: 'endpoint',
        };
        beforeEach(() => {
            component.formGroup.patchValue(data);
            fixture.detectChanges();
        });

        it('should emit form value', () => {
            const spy = spyOn(dialogRef, 'close');
            component.onSave();
            fixture.detectChanges();

            expect(spy).toHaveBeenCalledWith(data);
        });
    });
});
