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

import { TestBed } from '@angular/core/testing';

import { MatDialog, MatDialogRef } from '@angular/material/dialog';
import { MockMatDialog } from '@testing/mock-mat-dialog';
import { MockMatDialogRef } from '@testing/mock-mat-dialog-ref';
import { ConfirmDialogService } from './confirm-dialog.service';

describe('ConfirmDialogService', () => {
    let service: ConfirmDialogService;
    let dialog: MockMatDialog;

    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [
                { provide: MatDialogRef, useClass: MockMatDialogRef },
                { provide: MatDialog, useClass: MockMatDialog },
            ]
        });
        service = TestBed.inject(ConfirmDialogService);
        // @ts-ignore
        dialog = TestBed.inject(MatDialog);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should open dialog on confirm', () => {
        const spy = spyOn(dialog, 'open').and.callThrough();
        service.confirm('title', 'message');

        expect(spy).toHaveBeenCalled();
    });
});
